package application

import (
	"context"
	"testing"

	"pokertipssimulator/internal/adapter/repository"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
)

func setupActionTest(t *testing.T) (port.RoomRepository, *ActionUseCase) {
	t.Helper()
	db := repository.NewTestDB(t)
	repo := repository.NewSQLiteRoomRepository(db)
	uc := NewActionUseCase(repo)
	return repo, uc
}

func seedRoundRoom(t *testing.T, repo port.RoomRepository) *entity.Room {
	t.Helper()
	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels:       []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
				CurrentLevel: 0,
			},
		},
		Players: []entity.Player{
			{ID: "host", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p3",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "host", Bet: 5, TotalBet: 5},
				{PlayerID: "p2", Bet: 10, TotalBet: 10},
				{PlayerID: "p3", Bet: 0, TotalBet: 0},
			},
		},
	}
	ctx := context.Background()
	if err := repo.Create(ctx, room); err != nil {
		t.Fatalf("seed round room: %v", err)
	}
	return room
}

func TestProcessAction_Fold(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	room, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionFold, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(room.Round, "p3")
	if !ps.Folded {
		t.Error("expected p3 to be folded")
	}
	if !ps.HasActed {
		t.Error("expected HasActed to be true")
	}
}

func TestProcessAction_Check(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	room.Round.CurrentTurn = "p3"
	_ = repo.Update(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCheck, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p3")
	if !ps.HasActed {
		t.Error("expected HasActed to be true after check")
	}
}

func TestProcessAction_Check_InvalidWhenBetExists(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCheck, 0)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction when checking with bet on table, got %v", err)
	}
}

func TestProcessAction_Call(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	room, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCall, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p3 := room.FindPlayer("p3")
	ps := FindPlayerState(room.Round, "p3")
	if ps.Bet != 10 {
		t.Errorf("expected bet 10, got %d", ps.Bet)
	}
	if p3.Stack != 990 {
		t.Errorf("expected stack 990, got %d", p3.Stack)
	}
}

func TestProcessAction_Call_InvalidWhenNoBet(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCall, 0)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction when calling with no bet, got %v", err)
	}
}

func TestProcessAction_Bet(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	_ = repo.Update(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionBet, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p3")
	if ps.Bet != 20 {
		t.Errorf("expected bet 20, got %d", ps.Bet)
	}
	if result.Round.CurrentBet != 20 {
		t.Errorf("expected current bet 20, got %d", result.Round.CurrentBet)
	}

	p3 := result.FindPlayer("p3")
	if p3.Stack != 980 {
		t.Errorf("expected stack 980, got %d", p3.Stack)
	}
}

func TestProcessAction_Bet_InvalidWhenBetExists(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionBet, 20)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction when betting with existing bet, got %v", err)
	}
}

func TestProcessAction_Bet_MinimumBigBlind(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionBet, 5)
	if err != entity.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount for bet below BB, got %v", err)
	}
}

func TestProcessAction_Bet_InsufficientStack(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionBet, 2000)
	if err != entity.ErrInsufficientStack {
		t.Errorf("expected ErrInsufficientStack, got %v", err)
	}
}

func TestProcessAction_Raise(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionRaise, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p3")
	if ps.Bet != 20 {
		t.Errorf("expected bet 20, got %d", ps.Bet)
	}
	if result.Round.CurrentBet != 20 {
		t.Errorf("expected current bet 20, got %d", result.Round.CurrentBet)
	}

	p3 := result.FindPlayer("p3")
	if p3.Stack != 980 {
		t.Errorf("expected stack 980, got %d", p3.Stack)
	}
}

func TestProcessAction_Raise_InvalidWhenNoBet(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.CurrentBet = 0
	for i := range room.Round.PlayerStates {
		room.Round.PlayerStates[i].Bet = 0
	}
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionRaise, 20)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction when raising with no bet, got %v", err)
	}
}

func TestProcessAction_Raise_BelowMinimum(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionRaise, 15)
	if err != entity.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount for raise below min, got %v", err)
	}
}

func TestProcessAction_Raise_InsufficientStack(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.FindPlayer("p3").Stack = 10
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionRaise, 100)
	if err != entity.ErrInsufficientStack {
		t.Errorf("expected ErrInsufficientStack, got %v", err)
	}
}

func TestProcessAction_AllIn(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionAllIn, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p3")
	if !ps.AllIn {
		t.Error("expected AllIn to be true")
	}

	p3 := result.FindPlayer("p3")
	if p3.Stack != 0 {
		t.Errorf("expected stack 0, got %d", p3.Stack)
	}

	if ps.TotalBet != 1000 {
		t.Errorf("expected total bet 1000, got %d", ps.TotalBet)
	}
}

func TestProcessAction_NotYourTurn(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	_, err := uc.ProcessAction(ctx, "room-1", "host", entity.ActionFold, 0)
	if err != entity.ErrNotYourTurn {
		t.Errorf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestProcessAction_NoRound(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		HostPlayerID: "host",
		Players:      []entity.Player{{ID: "host"}},
	}
	_ = repo.Create(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "host", entity.ActionFold, 0)
	if err != entity.ErrGameNotStarted {
		t.Errorf("expected ErrGameNotStarted, got %v", err)
	}
}

func TestProcessAction_RoundComplete(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := seedRoundRoom(t, repo)
	room.Round.IsComplete = true
	_ = repo.Update(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionFold, 0)
	if err != entity.ErrRoundComplete {
		t.Errorf("expected ErrRoundComplete, got %v", err)
	}
}

func TestProcessAction_TurnAdvances(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, _ := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCall, 0)

	if result.Round.CurrentTurn == "p3" {
		t.Error("turn should advance after action")
	}
}

func TestProcessAction_AutoWin_AllFold(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p1",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 5, TotalBet: 5},
				{PlayerID: "p2", Bet: 10, TotalBet: 10},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p1", entity.ActionFold, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Round.IsComplete {
		t.Error("round should be complete when only one player remains")
	}
	if len(result.Round.Pots) != 1 {
		t.Fatalf("expected 1 pot, got %d", len(result.Round.Pots))
	}
	if result.Round.Pots[0].Amount != 15 {
		t.Errorf("expected pot 15, got %d", result.Round.Pots[0].Amount)
	}
	if result.Round.Pots[0].EligibleIDs[0] != "p2" {
		t.Errorf("expected p2 as winner, got %s", result.Round.Pots[0].EligibleIDs[0])
	}
}

func TestProcessAction_BettingRoundComplete(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetFlop,
			DealerSeat:  1,
			BigBlind:    10,
			CurrentBet:  0,
			MinRaise:    10,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 0, TotalBet: 10, HasActed: true},
				{PlayerID: "p2", Bet: 0, TotalBet: 10, HasActed: false},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, _ := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionCheck, 0)

	if result.Round.CurrentTurn != "" {
		t.Errorf("expected empty turn after betting complete, got %s", result.Round.CurrentTurn)
	}
}

func TestProcessAction_Bet_ResetsActedFlags(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 1000, Status: entity.PlayerStatusActive},
			{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetFlop,
			DealerSeat:  1,
			BigBlind:    10,
			CurrentBet:  0,
			MinRaise:    10,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", HasActed: false},
				{PlayerID: "p2", HasActed: false},
				{PlayerID: "p3", HasActed: false},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, _ := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionBet, 20)

	for _, ps := range result.Round.PlayerStates {
		if ps.PlayerID == "p2" {
			if !ps.HasActed {
				t.Error("bettor should have HasActed=true")
			}
		} else {
			if ps.HasActed {
				t.Errorf("player %s should have HasActed reset to false after bet", ps.PlayerID)
			}
		}
	}
}

func TestProcessAction_Call_AllInWhenShortStacked(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 5, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			BigBlind:    10,
			CurrentBet:  100,
			MinRaise:    100,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 100, TotalBet: 100, HasActed: true},
				{PlayerID: "p2", Bet: 0, TotalBet: 0},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, _ := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionCall, 0)

	p2 := result.FindPlayer("p2")
	ps := FindPlayerState(result.Round, "p2")
	if p2.Stack != 0 {
		t.Errorf("expected stack 0, got %d", p2.Stack)
	}
	if !ps.AllIn {
		t.Error("expected AllIn when stack depleted by call")
	}
	if ps.TotalBet != 5 {
		t.Errorf("expected total bet 5, got %d", ps.TotalBet)
	}
}

func TestProcessAction_InvalidActionType(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	_, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionType("invalid"), 0)
	if err != entity.ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction, got %v", err)
	}
}

func TestProcessAction_RecordsAction(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, _ := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCall, 0)

	if len(result.Round.Actions) != 1 {
		t.Fatalf("expected 1 action recorded, got %d", len(result.Round.Actions))
	}
	action := result.Round.Actions[0]
	if action.PlayerID != "p3" {
		t.Errorf("expected p3, got %s", action.PlayerID)
	}
	if action.Type != entity.ActionCall {
		t.Errorf("expected call, got %s", action.Type)
	}
	if action.Street != entity.StreetPreflop {
		t.Errorf("expected preflop, got %s", action.Street)
	}
}

func TestProcessAction_Raise_UpdatesMinRaise(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, _ := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionRaise, 30)

	if result.Round.MinRaise < 20 {
		t.Errorf("expected min raise >= 20, got %d", result.Round.MinRaise)
	}
}

func TestProcessAction_AllIn_AboveCurrentBet_ResetsFlags(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p3", Seat: 3, Stack: 500, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p3",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 5, TotalBet: 5, HasActed: true},
				{PlayerID: "p2", Bet: 10, TotalBet: 10, HasActed: true},
				{PlayerID: "p3", Bet: 0, TotalBet: 0},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionAllIn, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, ps := range result.Round.PlayerStates {
		if ps.PlayerID == "p3" {
			if !ps.AllIn {
				t.Error("p3 should be all-in")
			}
		} else {
			if ps.HasActed {
				t.Errorf("player %s should have HasActed reset after all-in raise", ps.PlayerID)
			}
		}
	}
}

func TestProcessAction_AllIn_BelowCurrentBet_NoReset(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 900, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 5, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  100,
			MinRaise:    100,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 100, TotalBet: 100, HasActed: true},
				{PlayerID: "p2", Bet: 0, TotalBet: 0},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionAllIn, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps2 := FindPlayerState(result.Round, "p2")
	if !ps2.AllIn {
		t.Error("p2 should be all-in")
	}
	if ps2.TotalBet != 5 {
		t.Errorf("expected total bet 5, got %d", ps2.TotalBet)
	}
	if result.Round.CurrentBet != 100 {
		t.Errorf("current bet should remain 100 when all-in below it, got %d", result.Round.CurrentBet)
	}
}

func TestProcessAction_PlayerNotFound(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "host", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			CurrentTurn: "ghost",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "host"},
			},
		},
	}
	_ = repo.Create(ctx, room)

	_, err := uc.ProcessAction(ctx, "room-1", "ghost", entity.ActionFold, 0)
	if err != entity.ErrPlayerNotFound {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestProcessAction_Bet_AllInWhenExactStack(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 1000, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 50, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetFlop,
			DealerSeat:  1,
			BigBlind:    10,
			CurrentBet:  0,
			MinRaise:    10,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1"},
				{PlayerID: "p2"},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionBet, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p2")
	if !ps.AllIn {
		t.Error("expected AllIn when bet equals entire stack")
	}
	p2 := result.FindPlayer("p2")
	if p2.Stack != 0 {
		t.Errorf("expected stack 0 after betting entire stack, got %d", p2.Stack)
	}
}

func TestProcessAction_Raise_AllInWhenExactStack(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "host",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 980, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 20, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p2",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 10, TotalBet: 10, HasActed: true},
				{PlayerID: "p2", Bet: 0, TotalBet: 0},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, err := uc.ProcessAction(ctx, "room-1", "p2", entity.ActionRaise, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ps := FindPlayerState(result.Round, "p2")
	if !ps.AllIn {
		t.Error("expected AllIn when raise uses entire stack")
	}
}

func TestProcessAction_Fold_PreservesOtherPlayerStates(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()
	seedRoundRoom(t, repo)

	result, _ := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionFold, 0)

	ps_host := FindPlayerState(result.Round, "host")
	ps_p2 := FindPlayerState(result.Round, "p2")

	if ps_host.Folded {
		t.Error("host should not be folded")
	}
	if ps_p2.Folded {
		t.Error("p2 should not be folded")
	}
}

func TestProcessAction_MultipleActionsSequence(t *testing.T) {
	repo, uc := setupActionTest(t)
	ctx := context.Background()

	room := &entity.Room{
		ID:           "room-1",
		Code:         "ABC123",
		Status:       entity.RoomStatusPlaying,
		HostPlayerID: "p1",
		Config: entity.RoomConfig{
			GameMode:      entity.GameModeCash,
			StartingStack: 1000,
			MaxPlayers:    10,
			BlindStructure: entity.BlindStructure{
				Levels: []entity.BlindLevel{{SmallBlind: 5, BigBlind: 10}},
			},
		},
		Players: []entity.Player{
			{ID: "p1", Seat: 1, Stack: 995, Status: entity.PlayerStatusActive},
			{ID: "p2", Seat: 2, Stack: 990, Status: entity.PlayerStatusActive},
			{ID: "p3", Seat: 3, Stack: 1000, Status: entity.PlayerStatusActive},
		},
		Round: &entity.Round{
			Number:      1,
			Street:      entity.StreetPreflop,
			DealerSeat:  1,
			SmallBlind:  5,
			BigBlind:    10,
			CurrentBet:  10,
			MinRaise:    10,
			CurrentTurn: "p3",
			PlayerStates: []entity.PlayerState{
				{PlayerID: "p1", Bet: 5, TotalBet: 5},
				{PlayerID: "p2", Bet: 10, TotalBet: 10},
				{PlayerID: "p3", Bet: 0, TotalBet: 0},
			},
		},
	}
	_ = repo.Create(ctx, room)

	result, _ := uc.ProcessAction(ctx, "room-1", "p3", entity.ActionCall, 0)

	if result.Round.CurrentTurn == "p3" {
		t.Error("turn should advance past p3 after call")
	}

	result, _ = uc.ProcessAction(ctx, "room-1", result.Round.CurrentTurn, entity.ActionCall, 0)

	if len(result.Round.Actions) != 2 {
		t.Errorf("expected 2 actions recorded, got %d", len(result.Round.Actions))
	}
}
