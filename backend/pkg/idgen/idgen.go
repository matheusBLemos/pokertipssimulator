package idgen

import (
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
)

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func NewID() string {
	return uuid.New().String()
}

func NewRoomCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars))))
		code[i] = codeChars[n.Int64()]
	}
	return string(code)
}
