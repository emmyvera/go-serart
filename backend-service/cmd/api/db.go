package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Initialize MongoDB connection
func initMongoDB(mongoURL string) (*mongo.Client, error) {
	// Create Connection
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// Connect
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Panicln("Error Connecting:", err)
		return nil, err
	}

	log.Println("Made Connection to mongoDB!!!", &c)

	return c, nil

}
