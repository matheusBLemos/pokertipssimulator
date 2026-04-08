package application

import (
	"context"
	"testing"

	"pokertipssimulator/internal/adapter/repository"
	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
)

func newGameTestDeps(t *testing.T) (port.RoomRepository, *GameUseCase) {
	t.Helper()
	db := repository.NewTestDB(t)
	repo := repository.NewSQLiteRoomRepository(db)
	uc := NewGameUseCase(repo)
	return repo, uc
}

func seedRoom(t *testing.T, repo port.RoomRepository, hostID string, players []entity.Player, opts ...func(*entity.Room)) *entity.Room {
	t.Helper()
	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusWaiting,
		HostPlayerID: hostID,
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{
					{SmallBlind: 5, BigBlind: 10, Ante: 0, Duration: 0},
				},
				CurrentLevel: 0,
			},
			ChipSet: entity.DefaultChipSet(),
		},
		Players: players,
	}
	for _, opt := range opts {
		opt(room)
	}
	ctx := context.Background()
	if err := repo.Create(ctx, room); err != nil {
		t.Fatalf("seed room: %v", err)
	}
	return room
}

func reseedRoom(t *testing.T, repo port.RoomRepository, room *entity.Room) {
	t.Helper()
	ctx := context.Background()
	if err := repo.Update(ctx, room); err != nil {
		t.Fatalf("reseed room: %v", err)
	}
}

func TestStartRound_Success(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Name: "Host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Name: "P2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	room, err := uc.StartRound(ctx, "room-1", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.Status != entity.RoomStatusPlaying {
		t.Errorf("expected playing, got %s", room.Status)
	}
	if room.Round == nil {
		t.Fatal("expected round to be set")
	}
	if room.Round.Street != entity.StreetPreflop {
		t.Errorf("expected preflop, got %s", room.Round.Street)
	}
	if room.RoundCount != 1 {
		t.Errorf("expected round count 1, got %d", room.RoundCount)
	}
}

func TestStartRound_NotHost(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.StartRound(ctx, "room-1", "p2")
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestStartRound_NotEnoughPlayers(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.StartRound(ctx, "room-1", "host")
	if err != entity.ErrNotEnoughPlayers {
		t.Errorf("expected ErrNotEnoughPlayers, got %v", err)
	}
}

func TestStartRound_UnseatedPlayersExcluded(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 0, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.StartRound(ctx, "room-1", "host")
	if err != entity.ErrNotEnoughPlayers {
		t.Errorf("expected ErrNotEnoughPlayers, got %v", err)
	}
}

func TestStartRound_BlindsPosted(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	room, _ := uc.StartRound(ctx, "room-1", "host")

	totalBlinds := 0
	for _, ps := range room.Round.PlayerStates {
		totalBlinds += ps.TotalBet
	}
	if totalBlinds != 15 {
		t.Errorf("expected total blinds 15 (5+10), got %d", totalBlinds)
	}
	if room.Round.CurrentBet != 10 {
		t.Errorf("expected current bet 10, got %d", room.Round.CurrentBet)
	}
}

func TestStartRound_HeadsUp_DealerIsSB(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	room, _ := uc.StartRound(ctx, "room-1", "host")

	dealerSeat := room.Round.DealerSeat
	var sbBet, bbBet int
	for _, ps := range room.Round.PlayerStates {
		p := room.FindPlayer(ps.PlayerID)
		if p.Seat == dealerSeat {
			sbBet = ps.TotalBet
		} else {
			bbBet = ps.TotalBet
		}
	}

	if sbBet != 5 {
		t.Errorf("expected SB bet 5, got %d", sbBet)
	}
	if bbBet != 10 {
		t.Errorf("expected BB bet 10, got %d", bbBet)
	}
}

func TestStartRound_FirstToAct_HeadsUp(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	room, _ := uc.StartRound(ctx, "room-1", "host")

	dealerSeat := room.Round.DealerSeat
	dealerPlayer := ""
	for _, p := range room.Players {
		if p.Seat == dealerSeat {
			dealerPlayer = p.ID
		}
	}

	if room.Round.CurrentTurn != dealerPlayer {
		t.Errorf("in heads-up preflop, dealer (SB) should act first; expected %s, got %s",
			dealerPlayer, room.Round.CurrentTurn)
	}
}

func TestStartRound_FirstToAct_MultiWay(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p4", Seat: 4, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "p1", players)

	room, _ := uc.StartRound(ctx, "room-1", "p1")

	dealerSeat := room.Round.DealerSeat

	seatToID := make(map[int]string)
	for _, p := range room.Players {
		seatToID[p.Seat] = p.ID
	}

	seats := []int{1, 2, 3, 4}
	dealerIdx := -1
	for i, s := range seats {
		if s == dealerSeat {
			dealerIdx = i
			break
		}
	}

	utgIdx := (dealerIdx + 3) % 4
	expectedFirstToAct := seatToID[seats[utgIdx]]

	if room.Round.CurrentTurn != expectedFirstToAct {
		t.Errorf("UTG should act first; expected %s, got %s", expectedFirstToAct, room.Round.CurrentTurn)
	}
}

func TestStartRound_EliminatedPlayersExcluded(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 0, Status: entity.PlayerStatusEliminated},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "host", players)

	room, err := uc.StartRound(ctx, "room-1", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(room.Round.PlayerStates) != 2 {
		t.Errorf("expected 2 player states (eliminated excluded), got %d", len(room.Round.PlayerStates))
	}

	for _, ps := range room.Round.PlayerStates {
		if ps.PlayerID == "p2" {
			t.Error("eliminated player should not be in round")
		}
	}
}

func TestAdvanceStreet_Success(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{
		Number:     1,
		Street:     entity.StreetPreflop,
		DealerSeat: 1,
		BigBlind:   10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "host", Bet: 10, TotalBet: 10, HasActed: true},
			{PlayerID: "p2", Bet: 10, TotalBet: 10, HasActed: true},
		},
	}
	reseedRoom(t, repo, room)

	updated, err := uc.AdvanceStreet(ctx, "room-1", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Round.Street != entity.StreetFlop {
		t.Errorf("expected flop, got %s", updated.Round.Street)
	}
	if updated.Round.CurrentBet != 0 {
		t.Errorf("expected currentBet 0, got %d", updated.Round.CurrentBet)
	}
	for _, ps := range updated.Round.PlayerStates {
		if ps.Bet != 0 {
			t.Error("expected bets reset to 0")
		}
		if ps.HasActed {
			t.Error("expected hasActed reset to false")
		}
	}
}

func TestAdvanceStreet_AllStreets(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
	}

	streets := []entity.Street{entity.StreetPreflop, entity.StreetFlop, entity.StreetTurn, entity.StreetRiver}
	expected := []entity.Street{entity.StreetFlop, entity.StreetTurn, entity.StreetRiver, entity.StreetShowdown}

	for i, street := range streets {
		// Delete and re-create room for each sub-test to avoid UNIQUE constraint violations
		ctx2 := context.Background()
		_ = repo.Delete(ctx2, "room-1")

		room := seedRoom(t, repo, "host", players)
		room.Round = &entity.Round{
			Number:     1,
			Street:     street,
			DealerSeat: 1,
			BigBlind:   10,
			PlayerStates: []entity.PlayerState{
				{PlayerID: "host", HasActed: true},
				{PlayerID: "p2", HasActed: true},
			},
		}
		reseedRoom(t, repo, room)

		updated, err := uc.AdvanceStreet(ctx, "room-1", "host")
		if err != nil {
			t.Fatalf("street %s: unexpected error: %v", street, err)
		}
		if updated.Round.Street != expected[i] {
			t.Errorf("from %s: expected %s, got %s", street, expected[i], updated.Round.Street)
		}
	}
}

func TestAdvanceStreet_ShowdownMarksComplete(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{
		Number:     1,
		Street:     entity.StreetRiver,
		DealerSeat: 1,
		BigBlind:   10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "host", HasActed: true},
			{PlayerID: "p2", HasActed: true},
		},
	}
	reseedRoom(t, repo, room)

	updated, _ := uc.AdvanceStreet(ctx, "room-1", "host")
	if !updated.Round.IsComplete {
		t.Error("expected round to be complete at showdown")
	}
	if updated.Round.CurrentTurn != "" {
		t.Error("expected no current turn at showdown")
	}
}

func TestAdvanceStreet_NotHost(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{Street: entity.StreetPreflop, BigBlind: 10, PlayerStates: []entity.PlayerState{}}
	reseedRoom(t, repo, room)

	_, err := uc.AdvanceStreet(ctx, "room-1", "p2")
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestAdvanceStreet_NoRound(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{{ID: "host", Seat: 1, Stack: 1000}}
	seedRoom(t, repo, "host", players)

	_, err := uc.AdvanceStreet(ctx, "room-1", "host")
	if err != entity.ErrGameNotStarted {
		t.Errorf("expected ErrGameNotStarted, got %v", err)
	}
}

func TestAdvanceStreet_InvalidFromShowdown(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{Street: entity.StreetShowdown, BigBlind: 10}
	reseedRoom(t, repo, room)

	_, err := uc.AdvanceStreet(ctx, "room-1", "host")
	if err != entity.ErrInvalidStreet {
		t.Errorf("expected ErrInvalidStreet, got %v", err)
	}
}

func TestSettleRound_Success(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 900, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Status = entity.RoomStatusPlaying
	room.Round = &entity.Round{
		Number:     1,
		Street:     entity.StreetShowdown,
		IsComplete: true,
		BigBlind:   10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "host", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
		},
	}
	reseedRoom(t, repo, room)

	updated, err := uc.SettleRound(ctx, "room-1", "host", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 0, PlayerIDs: []string{"host"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Round != nil {
		t.Error("expected round to be nil after settlement")
	}
	if updated.Status != entity.RoomStatusWaiting {
		t.Errorf("expected waiting, got %s", updated.Status)
	}

	host := updated.FindPlayer("host")
	if host.Stack != 1100 {
		t.Errorf("expected host stack 1100 (900+200), got %d", host.Stack)
	}
}

func TestSettleRound_SplitPot(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 900, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players)
	room.Round = &entity.Round{
		Street:   entity.StreetShowdown,
		BigBlind: 10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
		},
	}
	reseedRoom(t, repo, room)

	updated, _ := uc.SettleRound(ctx, "room-1", "p1", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 0, PlayerIDs: []string{"p1", "p2"}},
		},
	})

	p1 := updated.FindPlayer("p1")
	p2 := updated.FindPlayer("p2")
	if p1.Stack != 1000 || p2.Stack != 1000 {
		t.Errorf("expected even split: p1=%d p2=%d, expected 1000 each", p1.Stack, p2.Stack)
	}
}

func TestSettleRound_OddChipToFirstWinner(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 949, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 949, Status: entity.PlayerStatusActive},
		{ID: "p3", Seat: 3, Stack: 949, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players)
	room.Round = &entity.Round{
		Street:   entity.StreetShowdown,
		BigBlind: 10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 51},
			{PlayerID: "p2", TotalBet: 51},
			{PlayerID: "p3", TotalBet: 51},
		},
	}
	reseedRoom(t, repo, room)

	updated, _ := uc.SettleRound(ctx, "room-1", "p1", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 0, PlayerIDs: []string{"p1", "p2"}},
		},
	})

	p1 := updated.FindPlayer("p1")
	p2 := updated.FindPlayer("p2")
	totalWon := (p1.Stack - 949) + (p2.Stack - 949)
	if totalWon != 153 {
		t.Errorf("expected total winnings 153, got %d (p1=%d, p2=%d)", totalWon, p1.Stack, p2.Stack)
	}
	if p1.Stack != 949+77 {
		t.Errorf("first winner should get 77 (76+1 odd chip), p1 stack=%d", p1.Stack)
	}
	if p2.Stack != 949+76 {
		t.Errorf("second winner should get 76, p2 stack=%d", p2.Stack)
	}
}

func TestSettleRound_TournamentElimination(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 0, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players, func(r *entity.Room) {
		r.Config.GameMode = entity.GameModeTournament
	})
	room.Round = &entity.Round{
		Street:   entity.StreetShowdown,
		BigBlind: 10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
		},
	}
	reseedRoom(t, repo, room)

	updated, _ := uc.SettleRound(ctx, "room-1", "p1", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 0, PlayerIDs: []string{"p1"}},
		},
	})

	p2 := updated.FindPlayer("p2")
	if p2.Status != entity.PlayerStatusEliminated {
		t.Errorf("expected eliminated, got %s", p2.Status)
	}
}

func TestSettleRound_CashNoElimination(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 0, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players)
	room.Round = &entity.Round{
		Street:   entity.StreetShowdown,
		BigBlind: 10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
		},
	}
	reseedRoom(t, repo, room)

	updated, _ := uc.SettleRound(ctx, "room-1", "p1", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 0, PlayerIDs: []string{"p1"}},
		},
	})

	p2 := updated.FindPlayer("p2")
	if p2.Status == entity.PlayerStatusEliminated {
		t.Error("cash game should not eliminate players")
	}
}

func TestSettleRound_NotHost(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{
		BigBlind:     10,
		PlayerStates: []entity.PlayerState{{PlayerID: "host"}, {PlayerID: "p2"}},
	}
	reseedRoom(t, repo, room)

	_, err := uc.SettleRound(ctx, "room-1", "p2", dto.SettleRequest{})
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestPauseGame_Toggle(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Status = entity.RoomStatusPlaying
	reseedRoom(t, repo, room)

	updated, _ := uc.PauseGame(ctx, "room-1", "host")
	if updated.Status != entity.RoomStatusPaused {
		t.Errorf("expected paused, got %s", updated.Status)
	}

	updated, _ = uc.PauseGame(ctx, "room-1", "host")
	if updated.Status != entity.RoomStatusPlaying {
		t.Errorf("expected playing, got %s", updated.Status)
	}
}

func TestPauseGame_NotHost(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
		{ID: "p2", Seat: 2, Stack: 1000},
	}
	room := seedRoom(t, repo, "host", players)
	room.Status = entity.RoomStatusPlaying
	reseedRoom(t, repo, room)

	_, err := uc.PauseGame(ctx, "room-1", "p2")
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestRebuy_CashGame(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 500, Status: entity.PlayerStatusActive},
	}
	seedRoom(t, repo, "host", players)

	updated, err := uc.Rebuy(ctx, "room-1", "host", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	host := updated.FindPlayer("host")
	if host.Stack != 1500 {
		t.Errorf("expected 1500, got %d", host.Stack)
	}
}

func TestRebuy_DefaultAmount(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 0, Status: entity.PlayerStatusEliminated},
	}
	seedRoom(t, repo, "host", players)

	updated, _ := uc.Rebuy(ctx, "room-1", "host", 0)

	host := updated.FindPlayer("host")
	if host.Stack != 1000 {
		t.Errorf("expected default rebuy 1000, got %d", host.Stack)
	}
	if host.Status != entity.PlayerStatusWaiting {
		t.Errorf("expected waiting after rebuy, got %s", host.Status)
	}
}

func TestRebuy_TournamentDenied(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 500, Status: entity.PlayerStatusActive},
	}
	seedRoom(t, repo, "host", players, func(r *entity.Room) {
		r.Config.GameMode = entity.GameModeTournament
	})

	_, err := uc.Rebuy(ctx, "room-1", "host", 1000)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction, got %v", err)
	}
}

func TestRebuy_PlayerNotFound(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.Rebuy(ctx, "room-1", "nonexistent", 1000)
	if err != entity.ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestKickPlayer_Success(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
		{ID: "p2", Seat: 2, Stack: 1000},
		{ID: "p3", Seat: 3, Stack: 1000},
	}
	seedRoom(t, repo, "host", players)

	updated, err := uc.KickPlayer(ctx, "room-1", "host", "p2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updated.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(updated.Players))
	}
	if updated.FindPlayer("p2") != nil {
		t.Error("p2 should be removed")
	}
}

func TestKickPlayer_NotHost(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
		{ID: "p2", Seat: 2, Stack: 1000},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.KickPlayer(ctx, "room-1", "p2", "host")
	if err != entity.ErrNotHost {
		t.Errorf("expected ErrNotHost, got %v", err)
	}
}

func TestCalculatePots_SinglePot(t *testing.T) {
	round := &entity.Round{
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
			{PlayerID: "p3", TotalBet: 100},
		},
	}

	uc := &GameUseCase{}
	pots := uc.CalculatePots(round)
	if len(pots) != 1 {
		t.Fatalf("expected 1 pot, got %d", len(pots))
	}
	if pots[0].Amount != 300 {
		t.Errorf("expected 300, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleIDs) != 3 {
		t.Errorf("expected 3 eligible, got %d", len(pots[0].EligibleIDs))
	}
}

func TestCalculatePots_SidePots(t *testing.T) {
	round := &entity.Round{
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 50, AllIn: true},
			{PlayerID: "p2", TotalBet: 100},
			{PlayerID: "p3", TotalBet: 100},
		},
	}

	uc := &GameUseCase{}
	pots := uc.CalculatePots(round)
	if len(pots) != 2 {
		t.Fatalf("expected 2 pots, got %d", len(pots))
	}

	if pots[0].Amount != 150 {
		t.Errorf("main pot: expected 150 (50*3), got %d", pots[0].Amount)
	}
	if pots[1].Amount != 100 {
		t.Errorf("side pot: expected 100 (50*2), got %d", pots[1].Amount)
	}
}

func TestCalculatePots_FoldedPlayersIneligible(t *testing.T) {
	round := &entity.Round{
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100, Folded: true},
			{PlayerID: "p2", TotalBet: 100},
			{PlayerID: "p3", TotalBet: 100},
		},
	}

	uc := &GameUseCase{}
	pots := uc.CalculatePots(round)
	if len(pots) != 1 {
		t.Fatalf("expected 1 pot, got %d", len(pots))
	}

	for _, id := range pots[0].EligibleIDs {
		if id == "p1" {
			t.Error("folded player should not be eligible")
		}
	}
	if len(pots[0].EligibleIDs) != 2 {
		t.Errorf("expected 2 eligible, got %d", len(pots[0].EligibleIDs))
	}
}

func TestCalculatePots_AllZero(t *testing.T) {
	round := &entity.Round{
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 0},
			{PlayerID: "p2", TotalBet: 0},
		},
	}

	uc := &GameUseCase{}
	pots := uc.CalculatePots(round)
	if len(pots) != 1 {
		t.Fatalf("expected 1 empty pot, got %d", len(pots))
	}
	if pots[0].Amount != 0 {
		t.Errorf("expected 0, got %d", pots[0].Amount)
	}
}

func TestCalculatePots_MultipleSidePots(t *testing.T) {
	round := &entity.Round{
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 30, AllIn: true},
			{PlayerID: "p2", TotalBet: 80, AllIn: true},
			{PlayerID: "p3", TotalBet: 200},
			{PlayerID: "p4", TotalBet: 200},
		},
	}

	uc := &GameUseCase{}
	pots := uc.CalculatePots(round)
	if len(pots) != 3 {
		t.Fatalf("expected 3 pots, got %d", len(pots))
	}

	if pots[0].Amount != 120 {
		t.Errorf("main pot: expected 120 (30*4), got %d", pots[0].Amount)
	}
	if pots[1].Amount != 150 {
		t.Errorf("side pot 1: expected 150 (50*3), got %d", pots[1].Amount)
	}
	if pots[2].Amount != 240 {
		t.Errorf("side pot 2: expected 240 (120*2), got %d", pots[2].Amount)
	}
}

func TestNextStreet(t *testing.T) {
	tests := []struct {
		current  entity.Street
		expected entity.Street
	}{
		{entity.StreetPreflop, entity.StreetFlop},
		{entity.StreetFlop, entity.StreetTurn},
		{entity.StreetTurn, entity.StreetRiver},
		{entity.StreetRiver, entity.StreetShowdown},
		{entity.StreetShowdown, ""},
		{"invalid", ""},
	}

	uc := &GameUseCase{}
	for _, tt := range tests {
		result := uc.getNextStreet(tt.current)
		if result != tt.expected {
			t.Errorf("getNextStreet(%s): expected %s, got %s", tt.current, tt.expected, result)
		}
	}
}

func TestDealerRotation(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "p1", players)

	room1, _ := uc.StartRound(ctx, "room-1", "p1")
	firstDealer := room1.Round.DealerSeat

	room1.Status = entity.RoomStatusWaiting
	for i := range room1.Players {
		room1.Players[i].Status = entity.PlayerStatusWaiting
	}
	reseedRoom(t, repo, room1)

	room2, _ := uc.StartRound(ctx, "room-1", "p1")
	secondDealer := room2.Round.DealerSeat

	if secondDealer == firstDealer {
		t.Error("dealer should rotate between rounds")
	}
}

func TestStartRound_ShortStackedBlind(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 3, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	seedRoom(t, repo, "p1", players)

	room, err := uc.StartRound(ctx, "room-1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p1 := room.FindPlayer("p1")
	var p1State *entity.PlayerState
	for i := range room.Round.PlayerStates {
		if room.Round.PlayerStates[i].PlayerID == "p1" {
			p1State = &room.Round.PlayerStates[i]
			break
		}
	}

	if p1 == nil || p1State == nil {
		t.Fatal("p1 should be in round")
	}

	if p1.Stack < 0 {
		t.Errorf("stack should not go negative, got %d", p1.Stack)
	}
	if p1State.TotalBet > 3 {
		t.Errorf("short-stacked player should post at most 3, got %d", p1State.TotalBet)
	}
}

func TestSettleRound_NoRound(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
	}
	seedRoom(t, repo, "host", players)

	_, err := uc.SettleRound(ctx, "room-1", "host", dto.SettleRequest{})
	if err != entity.ErrGameNotStarted {
		t.Errorf("expected ErrGameNotStarted, got %v", err)
	}
}

func TestSettleRound_InvalidPotIndex(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 900, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players)
	room.Round = &entity.Round{
		Street:   entity.StreetShowdown,
		BigBlind: 10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", TotalBet: 100},
			{PlayerID: "p2", TotalBet: 100},
		},
	}
	reseedRoom(t, repo, room)

	updated, err := uc.SettleRound(ctx, "room-1", "p1", dto.SettleRequest{
		Winners: []dto.PotWinner{
			{PotIndex: 99, PlayerIDs: []string{"p1"}},
		},
	})
	if err != nil {
		t.Fatalf("invalid pot index should not error, got: %v", err)
	}

	p1 := updated.FindPlayer("p1")
	if p1.Stack != 900 {
		t.Errorf("player should not receive winnings for invalid pot index, stack=%d", p1.Stack)
	}
}

func TestAdvanceStreet_MinRaiseResets(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 900, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "host", players)
	room.Round = &entity.Round{
		Number:     1,
		Street:     entity.StreetPreflop,
		DealerSeat: 1,
		BigBlind:   10,
		MinRaise:   50,
		CurrentBet: 50,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "host", Bet: 50, TotalBet: 50, HasActed: true},
			{PlayerID: "p2", Bet: 50, TotalBet: 50, HasActed: true},
		},
	}
	reseedRoom(t, repo, room)

	updated, err := uc.AdvanceStreet(ctx, "room-1", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Round.MinRaise != 10 {
		t.Errorf("expected min raise reset to big blind (10), got %d", updated.Round.MinRaise)
	}
}

func TestAdvanceStreet_PostflopFirstToActAfterDealer(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
		{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
		{ID: "p3", Seat: 3, Stack: 990, Status: entity.PlayerStatusActive},
	}
	room := seedRoom(t, repo, "p1", players)
	room.Round = &entity.Round{
		Number:     1,
		Street:     entity.StreetPreflop,
		DealerSeat: 1,
		BigBlind:   10,
		PlayerStates: []entity.PlayerState{
			{PlayerID: "p1", HasActed: true},
			{PlayerID: "p2", HasActed: true},
			{PlayerID: "p3", HasActed: true},
		},
	}
	reseedRoom(t, repo, room)

	updated, err := uc.AdvanceStreet(ctx, "room-1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Round.CurrentTurn != "p2" {
		t.Errorf("expected first to act postflop to be p2 (after dealer seat 1), got %s", updated.Round.CurrentTurn)
	}
}

func TestPauseGame_WaitingStatusUnchanged(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 1000},
	}
	seedRoom(t, repo, "host", players)

	updated, _ := uc.PauseGame(ctx, "room-1", "host")
	if updated.Status != entity.RoomStatusWaiting {
		t.Errorf("waiting room should remain waiting, got %s", updated.Status)
	}
}

func TestRebuy_RevivesEliminatedPlayer(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "host", Seat: 1, Stack: 0, Status: entity.PlayerStatusEliminated},
	}
	seedRoom(t, repo, "host", players)

	updated, err := uc.Rebuy(ctx, "room-1", "host", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	host := updated.FindPlayer("host")
	if host.Status != entity.PlayerStatusWaiting {
		t.Errorf("expected status waiting after rebuy, got %s", host.Status)
	}
	if host.Stack != 500 {
		t.Errorf("expected stack 500, got %d", host.Stack)
	}
}

func TestStartRound_DealerWrapsAround(t *testing.T) {
	repo, uc := newGameTestDeps(t)
	ctx := context.Background()

	players := []entity.Player{
		{ID: "p1", Seat: 1, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusWaiting},
		{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusWaiting},
	}
	room := seedRoom(t, repo, "p1", players)

	room.RoundCount = 1
	room.Round = &entity.Round{DealerSeat: 3}
	reseedRoom(t, repo, room)

	room.RoundCount = 1
	room.Round = &entity.Round{DealerSeat: 3}
	room.Status = entity.RoomStatusWaiting
	for i := range room.Players {
		room.Players[i].Status = entity.PlayerStatusWaiting
	}
	reseedRoom(t, repo, room)

	started, err := uc.StartRound(ctx, "room-1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if started.Round.DealerSeat != 1 {
		t.Errorf("expected dealer to wrap around to seat 1, got %d", started.Round.DealerSeat)
	}
}
