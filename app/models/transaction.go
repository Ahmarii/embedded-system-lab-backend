package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var TypeAmounts = map[string]float64{
	"A": 10.0,
	"B": 20.0,
	"C": 20.0,
}

type Transaction struct {
	UUID             uuid.UUID `bson:"_id,omitempty"`
	UserUUID         uuid.UUID `bson:"user_uuid,omitempty"` // Reference to User
	UserRFIDHashed   []byte    `bson:"user_rfid_hashed"`    // Store the hashed RFID for the user
	Type             string    `bson:"type"`
	Amount           float64   `bson:"amount"`
	RemainingBalance float64   `bson:"remaining_balance"`
	CreatedAt        time.Time `bson:"created_at,omitempty"`
	DeletedAt        time.Time `bson:"deleted_at,omitempty"`
	IsDeleted        bool      `bson:"is_deleted"`
}

func (t *Transaction) SetAmount() {
	t.Amount = TypeAmounts[t.Type]
}

func IsValidTransactionType(transactionType string) bool {
	_, exists := TypeAmounts[transactionType]
	return exists
}

func NewTransaction(userUUID uuid.UUID, userRFIDHashed []byte, transactionType string, remainingBalance float64) (*Transaction, error) {
	if _, exists := TypeAmounts[transactionType]; !exists {
		return nil, fmt.Errorf("invalid transaction type: %s", transactionType)
	}

	// Create the transaction
	t := &Transaction{
		UUID:             uuid.New(),
		UserUUID:         userUUID,
		UserRFIDHashed:   userRFIDHashed,
		Type:             transactionType,
		RemainingBalance: remainingBalance,
		CreatedAt:        time.Now(),
		IsDeleted:        false,
	}

	t.SetAmount()

	return t, nil
}
