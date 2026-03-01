package idgen

import (
	"crypto/rand"
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func NewID() string {
	return primitive.NewObjectID().Hex()
}

func NewRoomCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars))))
		code[i] = codeChars[n.Int64()]
	}
	return string(code)
}
