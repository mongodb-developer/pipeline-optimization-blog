package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"pipeline_blog/appconfig"
	"strings"
	"time"

	"log"

	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MasterWG sync.WaitGroup

type TestResult struct {
	TestName        string           `bson:"TestName"`
	StartTime       time.Time        `bson:"StartTime"`
	EndTime         time.Time        `bson:"EndTime"`
	Duration        int              `bson:"Duration"`
	InstanceResults []InstanceResult `bson:"InstanceResults"`
	InstanceAverage int              `bson:"InstanceAverage"`
}

type InstanceResult struct {
	StartTime     time.Time `bson:"StartTime"`
	EndTime       time.Time `bson:"EndTime"`
	Duration      int       `bson:"Duration"`
	ConnectionNum int       `bson:"ConnectionNum"`
	RoutineNum    int       `bson:"RoutineNum"`
	City          string    `bson:"City"`
	DeviceName    string    `bson:"DeviceName"`
}

func CreateIndex(coll *mongo.Collection, indexModel mongo.IndexModel, wg *sync.WaitGroup) {

	defer wg.Done()
	name, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Name of Index Created: " + name)

}

func GetMongoDatabase(mongoDBURI, mongoDBName string) *mongo.Database {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoDBURI).SetTLSConfig(&tls.Config{}).SetAppName("InsuranceLoader"))
	if err != nil {
		log.Fatal(err)
	}
	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	return client.Database(mongoDBName)

}

func SeedCollection(mdb *mongo.Database, wg *sync.WaitGroup, collName string) {

	defer wg.Done()
	filter := bson.D{}
	if collName == "Profiles" {
		pattern := `^[A-Za-z]`
		filter = bson.D{
			{"contact.address.city", bson.D{{"$regex", pattern}}},
		}
	} else if collName == "Devices" {
		pattern := `^[0-9a-f]`
		filter = bson.D{
			{"deviceSN", bson.D{{"$regex", pattern}}},
		}
	} else if collName == "Mappings" {
		pattern := `^[A-Za-z]`
		filter = bson.D{
			{"profileID", bson.D{{"$regex", pattern}}},
		}
	}
	cursor, err := mdb.Collection(collName).Find(context.TODO(), filter)
	if err != nil {
		log.Fatalf("Failed to seed cach (%s): %v", collName, err)
	}
	defer cursor.Close(context.TODO())
	c := 0
	for cursor.Next(context.TODO()) {
		var seedDoc interface{}
		err = cursor.Decode(&seedDoc)
		if err != nil {
			log.Fatalf("Failed to decode seeding doc: %v", err)
		}
		c++
	}
	log.Printf("Decoded %d docs seeding %s collection", c, collName)
}

func HideIndex(indexName, collectionName string, db *mongo.Database, hide bool) {

	// Hide the index
	collModCommand := bson.D{
		{"collMod", collectionName},
		{"index", bson.D{
			{"name", indexName},
			{"hidden", hide},
		}},
	}
	res := db.RunCommand(context.TODO(), collModCommand)
	if res.Err() != nil {
		log.Fatal(res.Err())
	}
}

func CreateResultDoc(mdb *mongo.Database, testName string) {

	result := TestResult{
		TestName:        testName,
		InstanceResults: []InstanceResult{},
	}
	resultsColl := mdb.Collection(appconfig.ConfigData.ResultsColl)
	_, err := resultsColl.InsertOne(context.TODO(), result)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveDuration(startTime, endTime time.Time, mdb *mongo.Database, testName string) {

	resultsColl := mdb.Collection(appconfig.ConfigData.ResultsColl)

	//Save the execution duration back to MongoDB
	duration := int(endTime.UnixMilli() - startTime.UnixMilli())
	filter := bson.D{{"TestName", testName}}
	updates := bson.D{
		{"$set", bson.D{
			{"StartTime", startTime},
			{"EndTime", endTime},
			{"Duration", duration},
		}},
	}
	_, err := resultsColl.UpdateOne(context.TODO(), filter, updates)
	if err != nil {
		log.Fatal(err)
	}
}

func resolveSRV(uri string) ([]string, error) {
	// Parse the SRV URI to extract the host
	uriParts := strings.Split(uri, "://")
	if len(uriParts) != 2 {
		return nil, fmt.Errorf("invalid SRV URI")
	}
	host := strings.Split(uriParts[1], "/")[0]
	//remove credentials if they were provided.
	hostParts := strings.Split(host, "@")
	if len(hostParts) > 1 {
		host = hostParts[1]
	}
	_, srvRecords, err := net.LookupSRV("mongodb", "tcp", host)
	if err != nil {
		return nil, err
	}
	hosts := make([]string, len(srvRecords))
	for i, srv := range srvRecords {
		if len(hostParts) > 1 {
			hosts[i] = fmt.Sprintf("%s@%s:%d", hostParts[0], strings.TrimSuffix(srv.Target, "."), srv.Port)
		} else {
			hosts[i] = fmt.Sprintf("%s:%d", strings.TrimSuffix(srv.Target, "."), srv.Port)
		}
	}
	return hosts, nil
}

func GenerateDirectConnectionStrings(srvURI string) ([]string, error) {
	hosts, err := resolveSRV(srvURI)
	if err != nil {
		return nil, err
	}
	directConnectionStrings := make([]string, len(hosts))
	for i, host := range hosts {
		directConnectionStrings[i] = fmt.Sprintf("mongodb://%s/?directConnection=true", host)
	}
	return directConnectionStrings, nil
}
