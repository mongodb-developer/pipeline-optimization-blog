# pipeline-optimization-blog

This repository allows you to build a test data set and run the example aggregation pipelines discussed in the series of blog posts on MonogDB aggregation pipeline optimization, published in Medium at:

[Aggregation Optimization in MongoDB - A Case Study From The Field.](https://medium.com/@graeme_robinson/aggregation-optimization-in-mongodb-a-case-study-from-the-field-ca334d80718a)

## Program Description

This program builds a test data set in a MongoDB database for a fictional video streaming service. It builds two main collections; profiles and devices. Profiles represent users accessing the service, and devices are the devices from which users have accessed the service. A third collection - mappings - is an associative collection mapping a many-to-many relationship between profiles and devices.

After building the data set, the program then runs a series of five aggregation pipeline designs against the data set - an initial design, and then four incrementally improved designs, as discussed in the Medium articles.

Each pipeline design is executied a specified number of times and both the total execution to complete all iterations, and the average execution time for individual iterations saved in a specified collection. 

The number of connections established to MongoDB, and the number of GoRoutines (threads, essentially), used by each connection can be configured. This allows performance to be optimized for both te hardware form which the program is run, and the nodes on which MonogDB is running. 

When loading data, the program will divide the total number of profile documents to be created (and the corresponding device and mapping documents), among the available GoRoutines and load the data in parallel. In testing for the Medium articles, I found data load performance tended to by limited by available CPU on the server running the program. I found that 3 MongoDB connections each running 5 GoROutines (15 GoRoutines in total) tended to max out the available CPU on the AWS EC2 t2-xlarge instance I used to run the program. 

When executing the pipeline performance tests, the program will use direct connections to each of the nodes in the MongoDB replica set, allocating the specified number of MongDB connections to the available nodes on a round-robin basis. This allows the secondary nodes in the replica set to share the load associated with executing the pipeline iterations with the primary node, and is a common approach in read-intensive workloads (the default behaviour is all read traffic is directed to the primary node). In testing, I set the number of MongoDB connections used by the program to be equal to the number of nodes in the replica set (3), but you are welcome to experiment with more or less connections and GoRoutines per connection to see what works best for the cluster tier or hardware you are running MongoDB on.

### Profile fields and indexes

In the article series, the design of the profile documents is modified to add additional fields to support the later iterations of the pipeline design. These field are included by the program during the initial data build, but are ignored when executing the initial pipeline designs. Likewise, the indexes on the profiles collection are updated to support the later pipeline iterations. All of the indexes used are created during initial data build, but are set to be hidden. During pipeline execution, the index corresponding to that pipeline interation is made visible, and all other indexes remain hidden ensuring the pipeline execution can only use the relevant index. 

### Cache seeding

Before starting the execution pf pipline tests, and each time the visible index on the profiles collection is changed, the program runs a query against each of collections desinged to pull as much of the collection and corresponding index data as possible into the MongoDB cache. This can take several minutes to complete depending on the size of the data set created. 

## Running the code

To use the repository, you can either build the project from source:

`GOOS=linux GOARCH=amd64 go build -o pipeline-optimization`

Alternatively, you can run one of the pre-built executables in the "executables" folder.


## Program Configuration

The program reads three enviroment variables:

`MONGODB_DB_NAME`: the name of the MongoDB database you are connecting to

`MONGODB_CONFIG_COLL`: the name of a collection containing further configuration options

`MONGODB_URI`: the connection URI for your MongoDB instance.

These environment variables can either be read from the system directly or defined in a file named `.env` in the same directory as your executable.

On startup, the program connects to MongoDB using the connection URI and attempts to read a single document from the configuration collection in the specified database. The configuration document is expected to be in the following format:

```
{
  "Debug": false,
  "ResultsColl": "Results-t2xlarge-1m-M20",
  "Connections": 3,
  "GoRoutines": 5,
  "TestRuns": 300,
  "ReloadData": true,
  "RunTests": true,
  "Profiles": 1000005
}
```
`Debug`: a boolean value. Currently, it is ignored.

`ResultsColl`: a string value, this is the name of the collection the performance test results will be written to. The contents of this collection are replaced on each run so update this value to a new collection name if you wish to retain prior results. MongoDB will create the collection if it does not already exist.

`Connections`: an integer value, his is the number of connections to MongoDB the program will establish. Typically this is set to a multiple of the number of nodes in your replica set for optimal read performance, but bear in mind that during a data load, all connections will be made to the primary node.

`GoRoutines`: an integer value, this is the number of GoRoutines (essentially, threads), that will be run in parallel per connection i.e. the total number of GoRoutines will be this number multiplied by the connnections value. 

`TestRuns`: an integer value, This is the number of times each pipeline version will be run during a test cycle. The runs are split accross the available GoRoutines so this number should be divisible by (Connections * GoRoutines) (TODO - Add schema validation to enforce this)

`ReloadData`: a boolean value, this indicates whether the test data should be reloaded. If set to true, all data in the Profiles, Mappings, and Devices collections will be replaced. 

`RunTests`: a boolean value, this indicates whether the pipeline performance tests should be run. 

`Profiles`: an integer value, when reloading test data, this indicates the number of profile documents that should be created. The number of mapping and device documents will be proportional to this (approximately 3.4 device documents, and 5 mapping documents will be created for each profile document). The creation of documents will be split accross the available GoRoutines and executed in parallel, so this number should be divisible by (Connections * GoRoutines) (TODO - Add schema validation to enforce this)

When preparing to run the program, you will need to create the specified configuration collection and add this document to it. On doing so, MongoDB will automatically add an `_id` (unique identifier) value to the document.

## Results Output

A full run including both data load and pipeline test executions, will result in six documents being created in the specified results collection - 1 giving the elapsed time to complete the data load, and one for the execution of each of the five pipeline iterations. A results document has the following format:

```
{
  "_id": {
    "$oid": "678eb6da122dfb0f5228888d"
  },
  "TestName": "indexSort",
  "StartTime": {
    "$date": "2025-01-20T20:49:30.186Z"
  },
  "EndTime": {
    "$date": "2025-01-20T20:49:30.841Z"
  },
  "Duration": 655,
  "InstanceResults": [
    {
      "StartTime": {
        "$date": "2025-01-20T20:49:30.186Z"
      },
      "EndTime": {
        "$date": "2025-01-20T20:49:30.200Z"
      },
      "Duration": 14,
      "ConnectionNum": 2,
      "RoutineNum": 3,
      "City": "Los Angeles",
      "DeviceName": "iPhone 16"
    },
    {
      "StartTime": {
        "$date": "2025-01-20T20:49:30.186Z"
      },
      "EndTime": {
        "$date": "2025-01-20T20:49:30.200Z"
      },
      "Duration": 14,
      "ConnectionNum": 2,
      "RoutineNum": 1,
      "City": "Midland",
      "DeviceName": "OnePlus TV"
    },
    {
      "StartTime": {
        "$date": "2025-01-20T20:49:30.186Z"
      },
      "EndTime": {
        "$date": "2025-01-20T20:49:30.200Z"
      },
      "Duration": 14,
      "ConnectionNum": 3,
      "RoutineNum": 5,
      "City": "Frisco",
      "DeviceName": "iPhone 15"
    },
    ...
  ],
  "InstanceAverage": 14.36,
  "ExplainPlan": {...}
}
```
`_id` is the MongoDB allocated uniqe identifier for the document.

`Testname` specifies the pipeline iteration this results document applies to. Value will be one of 'originalPipeline', 'noUnwinds', 'noMapping', 'dupllicateDeviceNames', or 'indexSort'.

`StartTime` and `EndTime` specify the start and end time of the full set of test iterations for this pipeline.

`Duration` is the time in milliseconds to complete all test iterations for this pipeline

`Instanceresults` is an array with one element for each test iteration. Each element includes the start and end time of that test, which connection and GoROutine ran the test, and the city and device name used by test (see the Meium articles for more details about the query being executed by the pipeline).

`Instance Average` gives the average time in milliseconds to complerte a single test iteration.

`ExplainPlan` contains an explain plan for one iteration of this pipeline. This can be useful for understanding the performance of individual stages in the pipeline and confirming indexes are bing used as expected.

## Article Test Parameters

For the testing described in the Medium articles, a test data set of 1 million profiles was created. This resulted in 3.4 million profile documents and 5 million mapping documents also being created. The program was run on an AWS EC2 t2-xlarge x86-64 instance running Amazon Linux. MongoDB was running on a MongoDB Atlas 3-Node AWS M20 cluster. Both the MongoDB cluster and the EC2 instance running the program were in us-west2 (Oregon) region. Three connections to MongoDB, each running five GoRoutines, were used.

An M20 cluster was selected based on the size of the data set used and the RAM available to be used by the MongoDB cache (MongoDB reserves approximately half of a server's memory for it's cache). If you increase or decrease the size of the data set you create, you may wish to use a different server tier if running in Atlas.
