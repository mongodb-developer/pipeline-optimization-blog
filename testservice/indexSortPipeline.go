package testservice

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getIndexSortPipeline(city, deviceName string) mongo.Pipeline {

	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match",
				bson.D{
					{"contact.address.city", city},
					{"devices.deviceName", deviceName},
				},
			},
		},
		bson.D{{"$skip", 0}},
		bson.D{{"$limit", 10}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "Devices"},
					{"localField", "devices.deviceSN"},
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
	}
	return pipeline

}
