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

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) (*UserRepository, error) {
	repo := &UserRepository{collection: collection}
	if err := repo.ensureIndexes(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *UserRepository) CreateByRFID(rfid string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.NewUser()
	user.RFID_Hashed = models.HashRFID(rfid)

	if _, err := r.collection.InsertOne(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByRFID(rfid string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{
		"rfid_hashed": models.HashRFID(rfid),
		"is_deleted":  false,
	}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(userID uuid.UUID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{
		"_id":        userID,
		"is_deleted": false,
	}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"is_deleted": false,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) UpdateAmountByRFID(rfid string, amount float64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{
			"rfid_hashed": models.HashRFID(rfid),
			"is_deleted":  false,
		},
		bson.M{
			"$set": bson.M{
				"amount": amount,
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var user models.User
	if err := result.Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateAmountByID(userID uuid.UUID, amount float64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{
			"_id":        userID,
			"is_deleted": false,
		},
		bson.M{
			"$set": bson.M{
				"amount": amount,
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var user models.User
	if err := result.Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) DeleteByRFID(rfid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"rfid_hashed": models.HashRFID(rfid),
			"is_deleted":  false,
		},
		bson.M{
			"$set": bson.M{
				"is_deleted": true,
				"delete_at":  time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "rfid_hashed", Value: 1},
		},
		Options: options.Index().
			SetName("unique_active_rfid_hashed").
			SetUnique(true).
			SetPartialFilterExpression(bson.M{
				"is_deleted": false,
			}),
	})

	return err
}
