package entity

import (
	"errors"
	"testing"
)

func TestDomainErrorsAreSentinels(t *testing.T) {
	sentinels := []error{
		ErrRoomNotFound,
		ErrPlayerNotFound,
		ErrRoomFull,
		ErrSeatTaken,
		ErrInvalidAction,
		ErrNotYourTurn,
		ErrInsufficientStack,
		ErrGameInProgress,
		ErrGameNotStarted,
		ErrNotHost,
		ErrInvalidAmount,
		ErrNotEnoughPlayers,
		ErrAlreadyJoined,
		ErrInvalidCode,
		ErrRoundComplete,
		ErrInvalidStreet,
	}

	for _, err := range sentinels {
		if err == nil {
			t.Error("sentinel error should not be nil")
		}
		if err.Error() == "" {
			t.Error("sentinel error should have a message")
		}
	}

	if errors.Is(ErrRoomNotFound, ErrPlayerNotFound) {
		t.Error("different sentinel errors should not match")
	}
}

func TestDomainErrorWrapping(t *testing.T) {
	if !errors.Is(ErrRoomNotFound, ErrRoomNotFound) {
		t.Error("error should match itself")
	}
}
