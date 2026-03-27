package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultMongoURI              = "mongodb://admin:admin@localhost:27017"
	defaultMongoDatabase         = "embedded_lab"
	defaultUserCollection        = "users"
	defaultTransactionCollection = "transactions"
)

func ConnectMongo() (*mongo.Client, *mongo.Collection, *mongo.Collection, error) {
	uri := getEnv("MONGO_URI", defaultMongoURI)
	databaseName := getEnv("MONGO_DATABASE", defaultMongoDatabase)
	userCollectionName := getEnv("MONGO_USER_COLLECTION", defaultUserCollection)
	transactionCollectionName := getEnv("MONGO_TRANSACTION_COLLECTION", defaultTransactionCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, nil, err
	}

	database := client.Database(databaseName)
	userCollection := database.Collection(userCollectionName)
	transactionCollection := database.Collection(transactionCollectionName)

	return client, userCollection, transactionCollection, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
