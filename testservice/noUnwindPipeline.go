package testservice

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getNoUnwindPipeline(city, deviceName string) mongo.Pipeline {

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
		bson.D{
			{"$set",
				bson.D{
					{"_id", "$$REMOVE"},
					{"deviceSNs", "$$REMOVE"},
					{"devices", "$$REMOVE"},
					{"mappingData", "$$REMOVE"},
					{"customerType", "$$REMOVE"},
				},
			},
		},
		bson.D{{"$match", bson.D{{"deviceData", bson.D{{"$ne", bson.A{}}}}}}},
		bson.D{{"$sort", bson.D{{"profileID", 1}}}},
		bson.D{{"$skip", 0}},
		bson.D{{"$limit", 10}},
	}
	return pipeline
}
