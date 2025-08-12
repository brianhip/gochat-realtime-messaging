package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global MongoDB client and database instances
// These will be used throughout the application to interact with MongoDB
var Client *mongo.Client
var Database *mongo.Database

// InitMongoDB establishes a connection to MongoDB using the provided URI
// It also creates a database instance named "gochat" for our application
func InitMongoDB(uri string) error {
	// Create a context with timeout to prevent hanging connections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to connect to MongoDB using the provided URI
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// Ping the MongoDB server to verify the connection is working
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	// Store the client and database references globally for use throughout the app
	Client = client
	Database = client.Database("gochat")
	
	log.Println("Connected to MongoDB")
	return nil
}

// DisconnectMongoDB gracefully closes the MongoDB connection
// This should be called when the application shuts down
func DisconnectMongoDB() error {
	// Create a context with timeout for the disconnect operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Close the MongoDB connection
	return Client.Disconnect(ctx)
}