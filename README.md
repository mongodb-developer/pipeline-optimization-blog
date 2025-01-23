# pipeline-optimization-blog
Accompanying source code for Graeme's pipeline optimization blog

This repository allows you to build a test data set and run the example aggregation pipelines discussed in the series of blog posts on MonogDB aggregation pipeline optimization,published by Graeme Robinson at:

[text](https://medium.com/@graeme_robinson/aggregation-optimization-in-mongodb-a-case-study-from-the-field-ca334d80718a)

To use the repository, you can either build the project from source:

`GOOS=linux GOARCH=amd64 go build -o pipeline-optimization

Or you can run one of the pre-built executables in the "executables" folder.

The program reads three enviroment variables:

MONGODB_DB_NAME - the name of the MongoDB database you are connecting to
MONGODB_CONFIG_COLL - the name of a collection containing further configuration options
MONGODB_URI - the connection URI for your MongoDB instance.

These environment variables can either be read from the system directly or defined in a file named .env in the same directory as your executable.

On startup, the program connects to MongoDB using the connection URI and attempts to read a single document from the configuration collection in the specified database. The configuration document is expected to be in the following format:

```
{
  "_id": {
    "$oid": "67537112187ef90a3d531338"
  },
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
`_id`: is the MongoDB document identifier and can be ignored.

`Debug`: should be a boolean value. Currently, it is ignored.

`ResultsColl`: a string value, this is the name of the collection the performance test results will be written to. The contents of this collection are replaced on each run so update this value to a new collection name if you wish to retain prior results. MongoDB will create the collection if it does not already exist.

`Connections`: an integer value, his is the number of connections to MongoDB the program will establish. Typically this is set to a multiple of the number of nodes in your replica set for optimal read performance, but bear in mind that during a data load, all connections will be made to the primary node.

`GoRoutines`: an integer value, this is the number of GoRoutines (essentially, threads), that will be run in parallel per connection i.e. the total number of GoRoutines will be this number multiplied by the connnections value. 

`TestRuns`: an integer value, This is the number of times each pipeline version will be run during a test cycle. The runs are split accross the available GoRoutines so this number should be divisible by (Connections * GoRoutines) (TODO - Add schema validation to enforce this)

`ReloadData`: a boolean value, this indicates whether the test data should be reloaded. If set to true, all data in the Profiles, Mappings, and Devices collections will be replaced. 

`RunTests`: a boolean value, this indicates whether the pipeline performance tests should be run. 

`Profiles`: an integer value, when reloading test data, this indicates the number of profile documents that should be created. The number of mapping and device documents will be proportional to this (approximately 3.4 device documents, and 5 mapping documents will be created for each profile document). The creation of documents will be split accross the available GoRoutines and executed in parallel, so this number should be divisible by (Connections * GoRoutines) (TODO - Add schema validation to enforce this)






To help analyze the performance of the pipeline and the underlying data model, we built a test environment on a 3-node MongoDB Atlas M20 cluster. This environment contained 1 million profile documents, 2.7 million coverage documents, and over 5 million mapping documents linking profiles to devices. 

The performance of the pipeline was measured using an application written in Go using the native MongoDB Go driver. The application ran queries in a total of 15 concurrent goroutines (for our purposes, we can think of a goroutine as a thread), split evenly across three connections to the database - one connection to each of the three nodes in the MongoDB replica set. This configuration allowed secondary nodes in the replica set to share the search load with the primary node and is a common practice in read intensive workloads. 

The application was run on an AWS t2.xlarge x86-64 server running Amazon Linux. Both it and the MongoDB Atlas cluster were deployed in the US-West-1 region. Performance was measured end to end, i.e. it included overheads in sending the query from the application to the database, and then returning the results from the database to the application. Each modification to the data model and pipeline design was run 300 times, each run using a random combination of city and device name. By not simply repeating the same combination of city and device name we avoided potentially masking issues with MongoDB being under-provisioned with enough memory for effective caching. The mean execution time of each of the 300 iterations was recorded. 