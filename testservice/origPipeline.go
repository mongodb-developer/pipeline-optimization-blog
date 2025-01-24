package testservice

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getOrigPipeline(city, deviceName string) mongo.Pipeline {

	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"contact.address.city", city}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "Mappings"},
					{"localField", "profileID"},
					{"foreignField", "profileID"},
					{"as", "mappingData"},
				},
			},
		},
		bson.D{{"$unwind", "$mappingData"}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "Devices"},
					{"localField", "mappingData.deviceSN"},
					{"foreignField", "deviceSN"},
					{"pipeline",
						bson.A{
							bson.D{{"$match", bson.D{{"deviceName", deviceName}}}},
							bson.D{
								{"$set",
									bson.D{
										{"_id", "$$REMOVE"},
									},
								},
							},
						},
					},
					{"as", "deviceData"},
				},
			},
		},
		bson.D{{"$unwind", "$deviceData"}},
		bson.D{
			{"$group",
				bson.D{
					{"_id", "$profileID"},
					{"firstName", bson.D{{"$first", "$firstName"}}},
					{"lastName", bson.D{{"$first", "$lastName"}}},
					{"contact", bson.D{{"$first", "$contact"}}},
					{"ssn", bson.D{{"$first", "$SSN"}}},
					{"deviceData", bson.D{{"$push", "$deviceData"}}},
				},
			},
		},
		bson.D{
			{"$set",
				bson.D{
					{"profileID", "$_id"},
					{"_id", "$$REMOVE"},
				},
			},
		},
		bson.D{{"$sort", bson.D{{"profileID", 1}}}},
		bson.D{{"$skip", 0}},
		bson.D{{"$limit", 10}},
	}
	return pipeline
}
