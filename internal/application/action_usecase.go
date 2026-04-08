package application

import (
	"context"
	"sort"

	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
)

type ActionUseCase struct {
	repo port.RoomRepository
}

func NewActionUseCase(repo port.RoomRepository) *ActionUseCase {
	return &ActionUseCase{repo: repo}
}

func (uc *ActionUseCase) ProcessAction(ctx context.Context, roomID, playerID string, actionType entity.ActionType, amount int) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Round == nil {
		return nil, entity.ErrGameNotStarted
	}

	round := room.Round
	if round.IsComplete {
		return nil, entity.ErrRoundComplete
	}

	if round.CurrentTurn != playerID {
		return nil, entity.ErrNotYourTurn
	}

	player := room.FindPlayer(playerID)
	if player == nil {
		return nil, entity.ErrPlayerNotFound
	}

	ps := FindPlayerState(round, playerID)
	if ps == nil {
		return nil, entity.ErrPlayerNotFound
	}

	if err := uc.validateAction(round, player, ps, actionType, amount); err != nil {
		return nil, err
	}

	switch actionType {
	case entity.ActionFold:
		ps.Folded = true
		ps.HasActed = true

	case entity.ActionCheck:
		ps.HasActed = true

	case entity.ActionCall:
		callAmount := round.CurrentBet - ps.Bet
		if callAmount > player.Stack {
			callAmount = player.Stack
		}
		player.Stack -= callAmount
		ps.Bet += callAmount
		ps.TotalBet += callAmount
		ps.HasActed = true
		if player.Stack == 0 {
			ps.AllIn = true
		}

	case entity.ActionBet:
		player.Stack -= amount
		ps.Bet = amount
		ps.TotalBet += amount
		ps.HasActed = true
		round.CurrentBet = amount
		round.MinRaise = amount
		if player.Stack == 0 {
			ps.AllIn = true
		}
		uc.resetActedFlags(round, playerID)

	case entity.ActionRaise:
		raiseAmount := amount - ps.Bet
		player.Stack -= raiseAmount
		raiseDiff := amount - round.CurrentBet
		if raiseDiff > round.MinRaise {
			round.MinRaise = raiseDiff
		}
		ps.Bet = amount
		ps.TotalBet += raiseAmount
		ps.HasActed = true
		round.CurrentBet = amount
		if player.Stack == 0 {
			ps.AllIn = true
		}
		uc.resetActedFlags(round, playerID)

	case entity.ActionAllIn:
		allInAmount := player.Stack
		ps.Bet += allInAmount
		ps.TotalBet += allInAmount
		player.Stack = 0
		ps.AllIn = true
		ps.HasActed = true
		if ps.Bet > round.CurrentBet {
			raiseDiff := ps.Bet - round.CurrentBet
			if raiseDiff >= round.MinRaise {
				round.MinRaise = raiseDiff
			}
			round.CurrentBet = ps.Bet
			uc.resetActedFlags(round, playerID)
		}
	}

	round.Actions = append(round.Actions, entity.Action{
		PlayerID: playerID,
		Type:     actionType,
		Amount:   amount,
		Street:   round.Street,
	})

	uc.advanceTurn(room, round)

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *ActionUseCase) validateAction(round *entity.Round, player *entity.Player, ps *entity.PlayerState, actionType entity.ActionType, amount int) error {
	switch actionType {
	case entity.ActionFold:
		return nil

	case entity.ActionCheck:
		if round.CurrentBet > ps.Bet {
			return entity.ErrInvalidAction
		}
		return nil

	case entity.ActionCall:
		if round.CurrentBet <= ps.Bet {
			return entity.ErrInvalidAction
		}
		return nil

	case entity.ActionBet:
		if round.CurrentBet > 0 {
			return entity.ErrInvalidAction
		}
		if amount < round.BigBlind && amount < player.Stack {
			return entity.ErrInvalidAmount
		}
		if amount > player.Stack {
			return entity.ErrInsufficientStack
		}
		return nil

	case entity.ActionRaise:
		if round.CurrentBet == 0 {
			return entity.ErrInvalidAction
		}
		minRaise := round.CurrentBet + round.MinRaise
		if amount < minRaise && amount < player.Stack+ps.Bet {
			return entity.ErrInvalidAmount
		}
		if amount-ps.Bet > player.Stack {
			return entity.ErrInsufficientStack
		}
		return nil

	case entity.ActionAllIn:
		return nil

	default:
		return entity.ErrInvalidAction
	}
}

func (uc *ActionUseCase) resetActedFlags(round *entity.Round, exceptPlayerID string) {
	for i := range round.PlayerStates {
		if round.PlayerStates[i].PlayerID != exceptPlayerID &&
			!round.PlayerStates[i].Folded &&
			!round.PlayerStates[i].AllIn {
			round.PlayerStates[i].HasActed = false
		}
	}
}

func (uc *ActionUseCase) advanceTurn(room *entity.Room, round *entity.Round) {
	activePlayers := room.ActivePlayers()
	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[i].Seat < activePlayers[j].Seat
	})

	nonFolded := 0
	var lastStanding string
	for _, ps := range round.PlayerStates {
		if !ps.Folded {
			nonFolded++
			lastStanding = ps.PlayerID
		}
	}
	if nonFolded <= 1 {
		round.IsComplete = true
		round.CurrentTurn = ""
		if lastStanding != "" {
			totalPot := 0
			for _, ps := range round.PlayerStates {
				totalPot += ps.TotalBet
			}
			round.Pots = []entity.Pot{
				{Amount: totalPot, EligibleIDs: []string{lastStanding}},
			}
		}
		return
	}

	allActed := true
	for _, ps := range round.PlayerStates {
		if !ps.Folded && !ps.AllIn && !ps.HasActed {
			allActed = false
			break
		}
		if !ps.Folded && !ps.AllIn && ps.Bet != round.CurrentBet {
			allActed = false
			break
		}
	}

	if allActed {
		round.CurrentTurn = ""
		return
	}

	currentIdx := -1
	for i, p := range activePlayers {
		if p.ID == round.CurrentTurn {
			currentIdx = i
			break
		}
	}

	for i := 1; i <= len(activePlayers); i++ {
		idx := (currentIdx + i) % len(activePlayers)
		ps := FindPlayerState(round, activePlayers[idx].ID)
		if ps != nil && !ps.Folded && !ps.AllIn && !ps.HasActed {
			round.CurrentTurn = activePlayers[idx].ID
			return
		}
	}

	round.CurrentTurn = ""
}
