package testservice

import (
	"context"
	"os"

	"log"

	"sync"
	"time"

	"math/rand"

	"pipeline_blog/appconfig"
	"pipeline_blog/common"
	"pipeline_blog/loaderservice"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunPerformanceTests() {

	mongoDBURI := os.Getenv("MONGODB_URI")
	//Connection to the primary node - use this to manipulate indexes etc.
	mdb := common.GetMongoDatabase(mongoDBURI, os.Getenv("MONGODB_DB_NAME"))
	//Direct connections to each node in the cluster. Use these to spread search load across the nodes.
	mongoDirectURIS, err := common.GenerateDirectConnectionStrings(mongoDBURI)
	if err != nil {
		log.Fatal(err)
	}
	mongoDBName := os.Getenv("MONGODB_DB_NAME")

	connectionCount := appconfig.ConfigData.Connections
	routineCount := appconfig.ConfigData.GoRoutines
	testRuns := appconfig.ConfigData.TestRuns

	//Seed the cache on each replica set:
	log.Print("Seeding cache on each replica set node")
	var seedConnections []*mongo.Database
	for _, node := range mongoDirectURIS {
		//We use direct connections to each node in the replica set in a round-robin
		//allocation. That should spread the load accross all the nodes in the replica set
		db := common.GetMongoDatabase(node, mongoDBName)
		seedConnections = append(seedConnections, db)
	}
	defer func() {
		for _, seedConn := range seedConnections {
			seedConn.Client().Disconnect(context.TODO())
		}
	}()
	common.MasterWG.Add(3 * len(seedConnections))
	for _, seedConn := range seedConnections {
		go common.SeedCollection(seedConn, &common.MasterWG, "Profiles")
		go common.SeedCollection(seedConn, &common.MasterWG, "Mappings")
		go common.SeedCollection(seedConn, &common.MasterWG, "Devices")
	}
	common.MasterWG.Wait()
	log.Print("Cache seeding complete")

	//Work out the number of testRuns to be executed by each connection.
	connectionRunCount := testRuns / connectionCount

	//Create the necessary number of Mongo Client / Database connections
	//Use direct connections so we can spread the read load accross the replica set

	var connections []*mongo.Database

	nodes := len(mongoDirectURIS)
	currNode := 0

	for i := 0; i < connectionCount; i++ {
		//We use direct connections to each node in the replica set in a round-robin
		//allocation. That should spread the load accross all the nodes in the replica set
		db := common.GetMongoDatabase(mongoDirectURIS[currNode], mongoDBName)
		connections = append(connections, db)
		currNode++
		currNode = currNode % nodes
	}
	defer func() {
		for i := 0; i < connectionCount; i++ {
			if err := connections[i].Client().Disconnect(context.TODO()); err != nil {
				log.Fatal(err)
			}
		}
	}()

	//Create an array of sub-wait groups - one for each MDB connection
	var wgs []*sync.WaitGroup
	for i := 0; i < connectionCount; i++ {
		var wg sync.WaitGroup
		wgs = append(wgs, &wg)
	}

	//Run Original Pipeline Tests
	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)
	//Unhide the index used by this pipeline (and make sure we hide it again once we're done)
	common.HideIndex("contact.address.city_1", "Profiles", mdb, false)
	defer common.HideIndex("contact.address.city_1", "Profiles", mdb, true)
	//Create the results document for this sequence of tests
	common.CreateResultDoc(mdb, "originalPipeline")
	//Start a new Go Routine for each MDB connection
	startTime := time.Now()
	for i := 0; i < connectionCount; i++ {
		go runTests(i, connections[i], mdb, wgs[i], routineCount, connectionRunCount, "originalPipeline")
	}
	common.MasterWG.Wait()
	endTime := time.Now()
	//Save the execution duration back to MongoDB
	common.SaveDuration(startTime, endTime, mdb, "originalPipeline")
	log.Print("Original Pipeline tests completed")

	//Run no-unwinds pipeline tests
	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)
	//Create the results document for this sequence of tests
	common.CreateResultDoc(mdb, "noUnwinds")
	//Start a new Go Routine for each MDB connection
	startTime = time.Now()
	for i := 0; i < connectionCount; i++ {
		go runTests(i, connections[i], mdb, wgs[i], routineCount, connectionRunCount, "noUnwinds")
	}
	common.MasterWG.Wait()
	endTime = time.Now()
	//Save the execution duration back to MongoDB
	common.SaveDuration(startTime, endTime, mdb, "noUnwinds")
	log.Print("No unwind tests completed")

	//Run no-mapping collection pipeline tests
	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)
	//Create the results document for this sequence of tests
	common.CreateResultDoc(mdb, "noMapping")
	//Start a new Go Routine for each MDB connection
	startTime = time.Now()
	for i := 0; i < connectionCount; i++ {
		go runTests(i, connections[i], mdb, wgs[i], routineCount, connectionRunCount, "noMapping")
	}
	common.MasterWG.Wait()
	endTime = time.Now()
	//Save the execution duration back to MongoDB
	common.SaveDuration(startTime, endTime, mdb, "noMapping")
	common.HideIndex("contact.address.city_1", "Profiles", mdb, true)
	log.Print("No mapping tests completed")

	//Run duplicate group_ids pipeline tests
	//Create the results document for this sequence of tests
	common.CreateResultDoc(mdb, "duplicateDeviceNames")
	//Unhide the index used by this pipeline (and make sure we hide it again once we're done)
	common.HideIndex("contact.address.city_1_devices.deviceName_1", "Profiles", mdb, false)
	defer common.HideIndex("contact.address.city_1_devices.deviceName_1", "Profiles", mdb, true)

	//Reseed the collections with the new index on the prifuiles collection active
	log.Print("Cache reseeding started")
	common.MasterWG.Add(3 * len(seedConnections))
	for _, seedConn := range seedConnections {
		go common.SeedCollection(seedConn, &common.MasterWG, "Profiles")
		go common.SeedCollection(seedConn, &common.MasterWG, "Mappings")
		go common.SeedCollection(seedConn, &common.MasterWG, "Devices")
	}
	common.MasterWG.Wait()
	log.Print("Cache reseeding complete")

	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)
	//Start a new Go Routine for each MDB connection
	startTime = time.Now()
	for i := 0; i < connectionCount; i++ {
		go runTests(i, connections[i], mdb, wgs[i], routineCount, connectionRunCount, "duplicateDeviceNames")
	}
	common.MasterWG.Wait()
	endTime = time.Now()
	//Save the execution duration back to MongoDB
	common.SaveDuration(startTime, endTime, mdb, "duplicateDeviceNames")
	common.HideIndex("contact.address.city_1_devices.deviceName_1", "Profiles", mdb, true)
	log.Print("Duplicate device name tests completed")

	//Run index sort pipeline tests
	//Create the results document for this sequence of tests
	common.CreateResultDoc(mdb, "indexSort")
	//Unhide the index used by this pipeline (and make sure we hide it again once we're done)
	common.HideIndex("contact.address.city_1_devices.deviceName_1_profileID_1", "Profiles", mdb, false)
	defer common.HideIndex("contact.address.city_1_devices.deviceName_1_profileID_1", "Profiles", mdb, true)
	//Reseed the collections with the new index on the prifuiles collection active
	log.Print("Cache reseeding started")
	common.MasterWG.Add(3 * len(seedConnections))
	for _, seedConn := range seedConnections {
		go common.SeedCollection(seedConn, &common.MasterWG, "Profiles")
		go common.SeedCollection(seedConn, &common.MasterWG, "Mappings")
		go common.SeedCollection(seedConn, &common.MasterWG, "Devices")
	}
	common.MasterWG.Wait()
	log.Print("Cache reseeding complete")

	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)
	//Start a new Go Routine for each MDB connection
	startTime = time.Now()
	for i := 0; i < connectionCount; i++ {
		go runTests(i, connections[i], mdb, wgs[i], routineCount, connectionRunCount, "indexSort")
	}
	common.MasterWG.Wait()
	endTime = time.Now()
	//Save the execution duration back to MongoDB
	common.SaveDuration(startTime, endTime, mdb, "indexSort")
	common.HideIndex("contact.address.city_1_devices.deviceName_1_profileID_1", "Profiles", mdb, true)
	log.Print("Index sort tests completed")

}

func runTests(connectionNum int, mdbread, mdbwrite *mongo.Database, wg *sync.WaitGroup, goRoutines, runCount int, testName string) {

	defer common.MasterWG.Done()

	//Work out the number of members to be loaded by each goRoutine.
	routineRunCount := runCount / goRoutines

	//Seed cache:

	//Initialize the wait group
	wg.Add(goRoutines)

	for i := 0; i < goRoutines; i++ {
		go runPipeline(connectionNum, i, mdbread, mdbwrite, wg, routineRunCount, testName)
	}
	wg.Wait()

}

func runPipeline(connectionNum, routineNum int, mdbread, mdbwrite *mongo.Database, wg *sync.WaitGroup, runCount int, testName string) {

	defer wg.Done()
	profileColl := mdbread.Collection("Profiles")

	for x := 0; x < runCount; x++ {

		cityState := loaderservice.RandomCityState()
		deviceName := loaderservice.RandomDeviceName(rand.Intn(2) == 1)
		city := cityState["city"]

		var pipeline mongo.Pipeline = nil
		switch testName {
		case "originalPipeline":
			pipeline = getOrigPipeline(city, deviceName)
		case "noUnwinds":
			pipeline = getNoUnwindPipeline(city, deviceName)
		case "noMapping":
			pipeline = getNoMappingsPipeline(city, deviceName)
		case "duplicateDeviceNames":
			pipeline = getDuplicateDeviceNamesPipeline(city, deviceName)
		case "indexSort":
			pipeline = getIndexSortPipeline(city, deviceName)
		}
		startTime := time.Now()
		// Run the aggregation
		if pipeline != nil {
			cursor, err := profileColl.Aggregate(context.TODO(), pipeline)
			if err != nil {
				log.Fatalf("Failed to run aggregation: %v", err)
			}
			defer cursor.Close(context.TODO())

			var memberDocs []interface{}
			err = cursor.All(context.TODO(), &memberDocs)
			if err != nil {
				log.Fatalf("Failed to decode aggregation result: %v", err)
			}
		}
		endTime := time.Now()

		//Save the execution duration back to MongoDB
		var result common.InstanceResult
		result.StartTime = startTime
		result.EndTime = endTime
		result.Duration = int(endTime.UnixMilli() - startTime.UnixMilli())
		result.ConnectionNum = connectionNum + 1
		result.RoutineNum = routineNum + 1
		result.City = city
		result.DeviceName = deviceName

		filter := bson.D{{"TestName", testName}}
		updates := bson.A{
			bson.D{
				{"$set", bson.D{
					{"InstanceResults", bson.D{
						{"$concatArrays", bson.A{
							"$InstanceResults", bson.A{result},
						}},
					}},
				}},
			},
			bson.D{
				{"$set", bson.D{
					{"InstanceAverage", bson.D{{"$avg", "$InstanceResults.Duration"}}},
				}},
			},
		}
		resultsColl := mdbwrite.Collection(appconfig.ConfigData.ResultsColl)
		_, err := resultsColl.UpdateOne(context.TODO(), filter, updates)
		if err != nil {
			log.Fatal(err)
		}

		//If this was the last run for connection / goroutine 1, rerun the query and get the explain for it.
		if connectionNum == 0 && routineNum == 0 && x == (runCount-1) && pipeline != nil {
			//Get the explain plan for the aggregation
			explainCommand := bson.D{
				{"explain", bson.D{
					{"aggregate", "Profiles"},
					{"pipeline", pipeline},
					{"cursor", bson.D{}},
				}},
				{"verbosity", "executionStats"}, // verbosity can be "queryPlanner", "executionStats", or "allPlansExecution"
			}
			//Run the explain command
			var explainResult bson.M
			err = mdbwrite.RunCommand(context.TODO(), explainCommand).Decode(&explainResult)
			if err != nil {
				log.Fatalf("Failed to get explain plan: %v", err)
			}
			//Add the explain plan to the results document
			updates := bson.D{
				{"$set", bson.D{{"ExplainPlan", explainResult}}},
			}
			resultsColl := mdbwrite.Collection(appconfig.ConfigData.ResultsColl)
			_, err = resultsColl.UpdateOne(context.TODO(), filter, updates)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
