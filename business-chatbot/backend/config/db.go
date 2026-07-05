package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

// ConnectMongo establishes the connection to MongoDB using MONGO_URI and
// selects the database named by MONGO_DB. Call this once at startup.
func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "business_chatbot"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping error: %v", err)
	}

	Client = client
	DB = client.Database(dbName)
	log.Printf("connected to MongoDB database: %s", dbName)
}

// Collection is a small helper so controllers/services don't need to
// repeat DB.Collection(...) everywhere.
func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}
