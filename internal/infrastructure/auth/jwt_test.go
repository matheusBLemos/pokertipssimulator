package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, err := svc.GenerateToken("room-1", "player-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	roomID, playerID, isHost, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if roomID != "room-1" {
		t.Errorf("expected room-1, got %s", roomID)
	}
	if playerID != "player-1" {
		t.Errorf("expected player-1, got %s", playerID)
	}
	if !isHost {
		t.Error("expected isHost to be true")
	}
}

func TestJWTService_NonHostToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, _ := svc.GenerateToken("room-1", "player-2", false)
	_, _, isHost, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isHost {
		t.Error("expected isHost to be false")
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	_, _, _, err := svc.ValidateToken("completely-invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1")
	svc2 := NewJWTService("secret-2")

	token, _ := svc1.GenerateToken("room-1", "player-1", true)
	_, _, _, err := svc2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	claims := jwt.MapClaims{
		"room_id":   "room-1",
		"player_id": "player-1",
		"is_host":   true,
		"exp":       time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret"))

	_, _, _, err := svc.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWTService_EmptySecret(t *testing.T) {
	svc := NewJWTService("")

	token, err := svc.GenerateToken("room-1", "player-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	roomID, playerID, _, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roomID != "room-1" || playerID != "player-1" {
		t.Error("claims should still work with empty secret")
	}
}

func TestJWTService_DifferentTokensForDifferentPlayers(t *testing.T) {
	svc := NewJWTService("test-secret")

	token1, _ := svc.GenerateToken("room-1", "player-1", true)
	token2, _ := svc.GenerateToken("room-1", "player-2", false)

	if token1 == token2 {
		t.Error("different players should get different tokens")
	}
}

func TestJWTService_TokenContainsAllClaims(t *testing.T) {
	svc := NewJWTService("test-secret")

	tokenStr, _ := svc.GenerateToken("room-42", "player-99", true)

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)

	if claims["room_id"] != "room-42" {
		t.Errorf("expected room-42, got %v", claims["room_id"])
	}
	if claims["player_id"] != "player-99" {
		t.Errorf("expected player-99, got %v", claims["player_id"])
	}
	if claims["is_host"] != true {
		t.Errorf("expected true, got %v", claims["is_host"])
	}
	if claims["exp"] == nil {
		t.Error("expected exp claim")
	}
}
