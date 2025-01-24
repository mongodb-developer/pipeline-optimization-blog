package main

import (
	"context"
	"log"
	"os"

	"pipeline_blog/appconfig"
	"pipeline_blog/common"

	"pipeline_blog/loaderservice"
	"pipeline_blog/testservice"

	"github.com/joho/godotenv"
)

func main() {

	//The following will load environment variables from a .env file in the application root folder if one exists.
	godotenv.Load()

	mongoDBURI := os.Getenv("MONGODB_URI")
	mongoDBName := os.Getenv("MONGODB_DB_NAME")
	mongoConfigColl := os.Getenv("MONGODB_CONFIG_COLL")

	mongoDB := common.GetMongoDatabase(mongoDBURI, mongoDBName)
	defer func() {
		if err := mongoDB.Client().Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	//Read the config data from MDB
	_, err := appconfig.ReadConfigMDB(mongoDB, mongoConfigColl)
	if err != nil {
		msg := "Streaming Service Pipeline tester encountered an unexpected result reading config data from MDB: " + err.Error()
		log.Fatal(msg)
	}

	//drop the results collection from prior runs
	mongoDB.Collection(appconfig.ConfigData.ResultsColl).Drop(context.TODO())

	if appconfig.ConfigData.ReloadData {
		loaderservice.LoadData()
	}
	if appconfig.ConfigData.RunTests {
		testservice.RunPerformanceTests()
	}
}
