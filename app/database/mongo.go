package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultMongoURI       = "mongodb://admin:admin@localhost:27017"
	defaultMongoDatabase  = "embedded_lab"
	defaultUserCollection = "users"
)

func ConnectMongo() (*mongo.Client, *mongo.Collection, error) {
	uri := getEnv("MONGO_URI", defaultMongoURI)
	databaseName := getEnv("MONGO_DATABASE", defaultMongoDatabase)
	collectionName := getEnv("MONGO_USER_COLLECTION", defaultUserCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, err
	}

	collection := client.Database(databaseName).Collection(collectionName)
	return client, collection, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
