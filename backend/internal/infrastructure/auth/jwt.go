package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) GenerateToken(roomID, playerID string, isHost bool) (string, error) {
	claims := jwt.MapClaims{
		"room_id":   roomID,
		"player_id": playerID,
		"is_host":   isHost,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenStr string) (roomID, playerID string, isHost bool, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return "", "", false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", false, jwt.ErrSignatureInvalid
	}

	roomID, _ = claims["room_id"].(string)
	playerID, _ = claims["player_id"].(string)
	isHost, _ = claims["is_host"].(bool)
	return roomID, playerID, isHost, nil
}
