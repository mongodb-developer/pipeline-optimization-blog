package appconfig

import (
	"context"
	"encoding/json"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AppConfig contains config settings
type AppConfig struct {
	Debug       bool   `bson:"Debug"`
	ResultsColl string `bson:"ResultsColl"`
	Connections int    `bson:"Connections"`
	GoRoutines  int    `bson:"GoRoutines"`
	Profiles    int    `bson:"Profiles"` //Must be divisible by (Connections * GoRoutines)
	TestRuns    int    `bson:"TestRuns"` //Must be divisible by (Connections * GoRoutines)
	ReloadData  bool   `bson:"ReloadData"`
	RunTests    bool   `bson:"RunTests"`
}

// ConfigData contains application configuration settings read from a JSON formatted file.
var ConfigData AppConfig

func ReadConfigMDB(mongoDB *mongo.Database, configColl string) (configjson string, err error) {

	//read the config data from MongoDB
	coll := mongoDB.Collection(configColl)
	filter := bson.D{{}}
	err = coll.FindOne(context.TODO(), filter).Decode(&ConfigData)
	if err != nil {
		log.Fatalf("Error reading config info from MongoDB: %s", err)
	}
	if jbytes, err := json.Marshal(ConfigData); err != nil {
		return "", err
	} else {
		return string(jbytes), err
	}
}
