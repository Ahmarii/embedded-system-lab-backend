package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/KaningNoppasin/embedded-system-lab-backend/app/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionRepository struct {
	collection *mongo.Collection
}

var ErrInsufficientBalance = errors.New("insufficient balance")

func NewTransactionRepository(collection *mongo.Collection) (*TransactionRepository, error) {
	repo := &TransactionRepository{collection: collection}
	if err := repo.ensureIndexes(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *TransactionRepository) Create(user *models.User, transactionType string) (*models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transaction, err := models.NewTransaction(user.UUID, user.RFID_Hashed, transactionType, user.Amount)
	if err != nil {
		return nil, err
	}

	if user.Amount < transaction.Amount {
		return nil, ErrInsufficientBalance
	}

	transaction.RemainingBalance = user.Amount - transaction.Amount

	if _, err := r.collection.InsertOne(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) GetAll() ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) GetByUserRFIDHashed(userRFIDHashed []byte) ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_rfid_hashed": userRFIDHashed,
		"is_deleted":       false,
	}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_uuid", Value: 1}},
			Options: options.Index().
				SetName("user_uuid_idx"),
		},
		{
			Keys: bson.D{{Key: "user_rfid_hashed", Value: 1}},
			Options: options.Index().
				SetName("user_rfid_hashed_idx"),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("created_at_desc_idx"),
		},
	})

	return err
}

func ParseUserID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
