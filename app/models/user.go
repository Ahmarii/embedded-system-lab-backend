package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID        uuid.UUID `bson:"_id,omitempty"`
	RFID_Hashed []byte    `bson:"rfid_hashed"`
	Amount      float64   `bson:"amount"`
	CreateAt    time.Time `bson:"create_at,omitempty"`
	DeleteAt    time.Time `bson:"delete_at,omitempty"`
	IsDeleted   bool      `bson:"is_deleted"`
}

func NewUser() *User {
	return &User{
		UUID:      uuid.New(),
		Amount:    0.0,
		CreateAt:  time.Now(),
		DeleteAt:  time.Time{},
		IsDeleted: false,
	}
}

func HashRFID(rfid string) []byte {
	h := sha256.New()
	h.Write([]byte(rfid))
	return h.Sum(nil)
}

func HashRFIDHex(rfid string) string {
	return hex.EncodeToString(HashRFID(rfid))
}
