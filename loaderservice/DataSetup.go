package loaderservice

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"

	"log"

	"sync"
	"time"

	"pipeline_blog/appconfig"
	"pipeline_blog/common"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func LoadData() {

	log.Print("Data Load Started")

	mongoDBURI := os.Getenv("MONGODB_URI")
	mongoDBName := os.Getenv("MONGODB_DB_NAME")

	connectionCount := appconfig.ConfigData.Connections
	routineCount := appconfig.ConfigData.GoRoutines
	profiles := appconfig.ConfigData.Profiles

	//Work out the number of profiles to be loaded by each connection.
	profilesCount := profiles / connectionCount

	//Create the necessary number of Mongo Client / Database connections
	var connections []*mongo.Database
	for i := 0; i < connectionCount; i++ {
		db := common.GetMongoDatabase(mongoDBURI, mongoDBName)
		connections = append(connections, db)
	}
	defer func() {
		for i := 0; i < connectionCount; i++ {
			if err := connections[i].Client().Disconnect(context.TODO()); err != nil {
				log.Fatal(err)
			}
		}
	}()

	//Initialize the master wait group
	common.MasterWG.Add(connectionCount)

	//Create an array of sub-wait groups - one for each MDB connection
	var wgs []*sync.WaitGroup
	for i := 0; i < connectionCount; i++ {
		var wg sync.WaitGroup
		wgs = append(wgs, &wg)
	}

	//Drop the existing collections
	coll := connections[0].Collection("Profiles")
	coll.Drop(context.TODO())
	coll = connections[0].Collection("Devices")
	coll.Drop(context.TODO())
	coll = connections[0].Collection("Mappings")
	coll.Drop(context.TODO())

	//Start a new Go Routine for each MDB connection
	startTime := time.Now()
	startID := 1
	for i := 0; i < connectionCount; i++ {
		go runDataLoads(connections[i], wgs[i], routineCount, startID, profilesCount)
		startID += profilesCount
	}
	common.MasterWG.Wait()
	endTime := time.Now()

	//Save the execution duration back to MongoDB
	var result common.TestResult
	result.TestName = "Pipeline Blog Data Load"
	result.StartTime = startTime
	result.EndTime = endTime
	result.Duration = int(endTime.UnixMilli() - startTime.UnixMilli())
	resultsColl := connections[0].Collection(appconfig.ConfigData.ResultsColl)
	_, err := resultsColl.InsertOne(context.TODO(), result)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Data Load Completed - creating Indexes")

	common.MasterWG.Add(1)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"contact.address.city", 1},
		},
		Options: options.Index().SetHidden(true),
	}
	go common.CreateIndex(connections[0].Collection("Profiles"), indexModel, &common.MasterWG)

	common.MasterWG.Add(1)
	indexModel = mongo.IndexModel{
		Keys: bson.D{
			{"contact.address.city", 1},
			{"devices.deviceName", 1},
		},
		Options: options.Index().SetHidden(true),
	}
	go common.CreateIndex(connections[0].Collection("Profiles"), indexModel, &common.MasterWG)

	common.MasterWG.Add(1)
	indexModel = mongo.IndexModel{
		Keys: bson.D{
			{"contact.address.city", 1},
			{"devices.deviceName", 1},
			{"profileID", 1},
		},
		Options: options.Index().SetHidden(true),
	}
	go common.CreateIndex(connections[0].Collection("Profiles"), indexModel, &common.MasterWG)

	common.MasterWG.Add(1)
	indexModel = mongo.IndexModel{
		Keys: bson.D{
			{"profileID", 1},
		},
	}
	go common.CreateIndex(connections[0].Collection("Mappings"), indexModel, &common.MasterWG)

	common.MasterWG.Add(1)
	indexModel = mongo.IndexModel{
		Keys: bson.D{
			{"deviceSN", 1},
			{"deviceName", 1},
		},
	}
	go common.CreateIndex(connections[0].Collection("Devices"), indexModel, &common.MasterWG)

	common.MasterWG.Wait()

}

func insertData(mdb *mongo.Database, wg *sync.WaitGroup, startProfile, endProfile int) {

	defer wg.Done()

	//Use Write Concern 1 for faster performance. Not recommended for a production system
	wc := writeconcern.W1()
	var collOpts options.CollectionOptions
	collOpts.WriteConcern = wc
	profileColl := mdb.Collection("Profiles", &collOpts)
	deviceColl := mdb.Collection("Devices", &collOpts)
	mappingsColl := mdb.Collection("Mappings", &collOpts)

	var insertOrderedOp options.InsertManyOptions

	var profileDocs []interface{}
	var deviceDocs []interface{}
	var mappingDocs []interface{}

	for x := startProfile; x < endProfile; {

		var deviceSNs []string
		var deviceNames []string

		famSize := rand.Intn(2) + 1
		lastname := RandomLastName()
		accountNum := randomAccountIDBase()
		deviceNum := rand.Intn(5) + 1

		//Create the family's shared devices.
		for i := 0; i < deviceNum; i++ {
			deviceJSON, deviceSN, deviceName := GenerateDevice(true)
			deviceSNs = append(deviceSNs, deviceSN)
			deviceNames = append(deviceNames, deviceName)
			var deviceBSON bson.D
			err := bson.UnmarshalExtJSON([]byte(deviceJSON), true, &deviceBSON)
			if err != nil {
				log.Fatalf("Failed to convert Device JSON to BSON: %v", err)
			}
			deviceDocs = append(deviceDocs, deviceBSON)
		}

		//Create Profiles
		var primaryage int
		address := randomAddress()
		for i := 1; i <= famSize; i++ {
			var profile Profile
			profileDeviceSNs := deviceSNs
			profileDeviceNames := deviceNames
			//Generate the profile's personal devices
			personDevicesCount := rand.Intn(5)
			for i := 0; i < personDevicesCount; i++ {
				deviceJSON, deviceSN, deviceName := GenerateDevice(false)
				profileDeviceSNs = append(profileDeviceSNs, deviceSN)
				profileDeviceNames = append(profileDeviceNames, deviceName)
				var deviceBSON bson.D
				err := bson.UnmarshalExtJSON([]byte(deviceJSON), true, &deviceBSON)
				if err != nil {
					log.Fatalf("Failed to convert Device JSON to BSON: %v", err)
				}
				deviceDocs = append(deviceDocs, deviceBSON)
			}
			if i == 1 {
				//Primary Profile
				profile, primaryage = GenerateProfile(lastname, strconv.Itoa(i), "P", accountNum, address, 0, profileDeviceSNs, profileDeviceNames)
			} else if i == 2 {
				//Spouse of primary
				var spouseage int
				profile, spouseage = GenerateProfile(lastname, strconv.Itoa(i), "S", accountNum, address, primaryage, profileDeviceSNs, profileDeviceNames)
				if spouseage < 64 && primaryage < 64 {
					//expand family to include possible children
					famSize = rand.Intn(5) + 2 //Between 2 and 6
				}
				if spouseage < primaryage {
					primaryage = spouseage //this makes sure any children we generate aren't too old for one or both parents
				}
			} else {
				//Child
				profile, _ = GenerateProfile(lastname, strconv.Itoa(i), "C", accountNum, address, primaryage, profileDeviceSNs, profileDeviceNames)
			}
			//var profileBSON bson.D
			//err := bson.UnmarshalExtJSON([]byte(profileJSON), true, &profileBSON)
			//if err != nil {
			//	log.Fatalf("Failed to convert Profile JSON String to BSON: %v", err)
			//}
			//profileDocs = append(profileDocs, profileBSON)
			profileDocs = append(profileDocs, profile)

			//Add mappings
			for g := 0; g < len(profileDeviceSNs); g++ {
				// Create a map representing the profile to device mapping
				mappingData := map[string]interface{}{
					"profileID": accountNum + "-" + strconv.Itoa(i),
					"deviceSN":  profileDeviceSNs[g],
				}
				// Convert the map to a JSON string
				mappingJSON, err := json.Marshal(mappingData)
				if err != nil {
					panic(err)
				}
				var mappingBSON bson.D
				err = bson.UnmarshalExtJSON([]byte(mappingJSON), true, &mappingBSON)
				if err != nil {
					log.Fatalf("Failed to convert Mapping JSON to BSON: %v", err)
				}
				mappingDocs = append(mappingDocs, mappingBSON)

			}

			x++
			if x >= endProfile {
				break
			}
		}

		if len(deviceDocs) >= 10000 || x >= endProfile {
			_, err := deviceColl.InsertMany(context.TODO(), deviceDocs, insertOrderedOp.SetOrdered(false))
			if err != nil {
				log.Print(err)
			}
			deviceDocs = nil
			if appconfig.ConfigData.Debug {
				log.Printf("Device document batch written")
			}
		}

		if len(profileDocs) >= 10000 || x >= endProfile {
			_, err := profileColl.InsertMany(context.TODO(), profileDocs, insertOrderedOp.SetOrdered(false))
			if err != nil {
				log.Print(err)
			}
			profileDocs = nil
			if appconfig.ConfigData.Debug {
				log.Printf("Profile document batch written")
			}
		}

		if len(mappingDocs) >= 10000 || x >= endProfile {
			_, err := mappingsColl.InsertMany(context.TODO(), mappingDocs, insertOrderedOp.SetOrdered(false))
			if err != nil {
				log.Print(err)
			}
			mappingDocs = nil
			if appconfig.ConfigData.Debug {
				log.Printf("Mapping document batch written")
			}
		}

	}
}

// Function to retrieve a field value from bson.D
func getFieldValue(doc bson.D, fieldName string) (interface{}, bool) {
	for _, elem := range doc {
		if elem.Key == fieldName {
			return elem.Value, true
		}
	}
	return nil, false // Field not found
}

func runDataLoads(mdb *mongo.Database, wg *sync.WaitGroup, goRoutines, startID, customerCount int) {

	defer common.MasterWG.Done()

	//Work out the number of members to be loaded by each goRoutine.
	routineCustomerCount := customerCount / goRoutines

	//Initialize the wait group
	wg.Add(goRoutines)

	for i := 0; i < goRoutines; i++ {
		go insertData(mdb, wg, startID, startID+routineCustomerCount)
		startID += routineCustomerCount
	}
	wg.Wait()

}
