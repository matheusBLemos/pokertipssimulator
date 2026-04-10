package entity

import (
	"testing"
)

func TestFindPlayer(t *testing.T) {
	room := &Room{
		Players: []Player{
			{ID: "p1", Name: "Alice", Seat: 1, Stack: 1000, Status: PlayerStatusActive},
			{ID: "p2", Name: "Bob", Seat: 2, Stack: 500, Status: PlayerStatusActive},
			{ID: "p3", Name: "Charlie", Seat: 0, Stack: 1000, Status: PlayerStatusWaiting},
		},
	}

	t.Run("finds existing player", func(t *testing.T) {
		p := room.FindPlayer("p2")
		if p == nil {
			t.Fatal("expected to find player p2")
		}
		if p.Name != "Bob" {
			t.Errorf("expected Bob, got %s", p.Name)
		}
	})

	t.Run("returns nil for non-existent player", func(t *testing.T) {
		p := room.FindPlayer("p999")
		if p != nil {
			t.Fatal("expected nil for non-existent player")
		}
	})

	t.Run("returns pointer to actual player allowing mutation", func(t *testing.T) {
		p := room.FindPlayer("p1")
		p.Stack = 2000
		if room.Players[0].Stack != 2000 {
			t.Error("mutation through pointer should reflect on original slice")
		}
		room.Players[0].Stack = 1000
	})
}

func TestFindPlayerState(t *testing.T) {
	room := &Room{
		Round: &Round{
			PlayerStates: []PlayerState{
				{PlayerID: "p1", Bet: 10, TotalBet: 10},
				{PlayerID: "p2", Bet: 20, TotalBet: 20},
			},
		},
	}

	t.Run("finds existing player state", func(t *testing.T) {
		ps := room.FindPlayerState("p1")
		if ps == nil {
			t.Fatal("expected to find player state for p1")
		}
		if ps.Bet != 10 {
			t.Errorf("expected Bet=10, got %d", ps.Bet)
		}
	})

	t.Run("returns nil for non-existent player state", func(t *testing.T) {
		ps := room.FindPlayerState("p999")
		if ps != nil {
			t.Fatal("expected nil for non-existent player state")
		}
	})

	t.Run("returns nil when no round", func(t *testing.T) {
		roomNoRound := &Room{}
		ps := roomNoRound.FindPlayerState("p1")
		if ps != nil {
			t.Fatal("expected nil when round is nil")
		}
	})
}

func TestActivePlayers(t *testing.T) {
	room := &Room{
		Players: []Player{
			{ID: "p1", Status: PlayerStatusActive},
			{ID: "p2", Status: PlayerStatusWaiting},
			{ID: "p3", Status: PlayerStatusActive},
			{ID: "p4", Status: PlayerStatusEliminated},
		},
	}

	active := room.ActivePlayers()
	if len(active) != 2 {
		t.Fatalf("expected 2 active players, got %d", len(active))
	}
	if active[0].ID != "p1" || active[1].ID != "p3" {
		t.Errorf("expected p1 and p3, got %s and %s", active[0].ID, active[1].ID)
	}
}

func TestActivePlayersEmpty(t *testing.T) {
	room := &Room{
		Players: []Player{
			{ID: "p1", Status: PlayerStatusWaiting},
			{ID: "p2", Status: PlayerStatusEliminated},
		},
	}

	active := room.ActivePlayers()
	if len(active) != 0 {
		t.Fatalf("expected 0 active players, got %d", len(active))
	}
}

func TestSeatedPlayers(t *testing.T) {
	room := &Room{
		Players: []Player{
			{ID: "p1", Seat: 1},
			{ID: "p2", Seat: 0},
			{ID: "p3", Seat: 3},
			{ID: "p4", Seat: 0},
		},
	}

	seated := room.SeatedPlayers()
	if len(seated) != 2 {
		t.Fatalf("expected 2 seated players, got %d", len(seated))
	}
	if seated[0].ID != "p1" || seated[1].ID != "p3" {
		t.Errorf("expected p1 and p3, got %s and %s", seated[0].ID, seated[1].ID)
	}
}

func TestSeatedPlayersEmpty(t *testing.T) {
	room := &Room{
		Players: []Player{
			{ID: "p1", Seat: 0},
			{ID: "p2", Seat: 0},
		},
	}

	seated := room.SeatedPlayers()
	if len(seated) != 0 {
		t.Fatalf("expected 0 seated players, got %d", len(seated))
	}
}
