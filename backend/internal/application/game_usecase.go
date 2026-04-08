package application

import (
	"context"
	"sort"

	"pokertipssimulator/internal/application/dto"
	"pokertipssimulator/internal/application/port"
	"pokertipssimulator/internal/domain/entity"
)

type GameUseCase struct {
	repo        port.RoomRepository
	broadcaster port.WSBroadcaster
}

func NewGameUseCase(repo port.RoomRepository, broadcaster port.WSBroadcaster) *GameUseCase {
	return &GameUseCase{repo: repo, broadcaster: broadcaster}
}

func (uc *GameUseCase) StartRound(ctx context.Context, roomID, playerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	seated := room.SeatedPlayers()
	if len(seated) < 2 {
		return nil, entity.ErrNotEnoughPlayers
	}

	for i := range room.Players {
		if room.Players[i].Seat > 0 && room.Players[i].Status != entity.PlayerStatusEliminated {
			room.Players[i].Status = entity.PlayerStatusActive
		}
	}

	room.RoundCount++
	room.Status = entity.RoomStatusPlaying

	activePlayers := room.ActivePlayers()
	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[i].Seat < activePlayers[j].Seat
	})

	dealerSeat := uc.nextDealer(room, activePlayers)
	blindLevel := room.Config.BlindStructure.Levels[room.Config.BlindStructure.CurrentLevel]

	var playerStates []entity.PlayerState
	for _, p := range activePlayers {
		playerStates = append(playerStates, entity.PlayerState{
			PlayerID: p.ID,
		})
	}

	round := &entity.Round{
		Number:       room.RoundCount,
		Street:       entity.StreetPreflop,
		DealerSeat:   dealerSeat,
		SmallBlind:   blindLevel.SmallBlind,
		BigBlind:     blindLevel.BigBlind,
		PlayerStates: playerStates,
		Pots:         []entity.Pot{},
		Actions:      []entity.Action{},
	}

	uc.postBlinds(room, round, activePlayers)
	uc.setFirstToAct(round, activePlayers)

	room.Round = round

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) AdvanceStreet(ctx context.Context, roomID, playerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	if room.Round == nil {
		return nil, entity.ErrGameNotStarted
	}

	round := room.Round
	nextStreet := nextStreet(round.Street)
	if nextStreet == "" {
		return nil, entity.ErrInvalidStreet
	}

	collectBets(round)

	round.Street = nextStreet
	round.CurrentBet = 0
	round.MinRaise = round.BigBlind

	for i := range round.PlayerStates {
		round.PlayerStates[i].Bet = 0
		round.PlayerStates[i].HasActed = false
	}

	if nextStreet == entity.StreetShowdown {
		round.IsComplete = true
		round.CurrentTurn = ""
	} else {
		activePlayers := room.ActivePlayers()
		sort.Slice(activePlayers, func(i, j int) bool {
			return activePlayers[i].Seat < activePlayers[j].Seat
		})
		setFirstToActPostflop(round, activePlayers)
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) SettleRound(ctx context.Context, roomID, playerID string, req dto.SettleRequest) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	if room.Round == nil {
		return nil, entity.ErrGameNotStarted
	}

	collectBets(room.Round)

	pots := CalculatePots(room.Round)
	room.Round.Pots = pots

	for _, w := range req.Winners {
		if w.PotIndex < 0 || w.PotIndex >= len(pots) {
			continue
		}
		pot := pots[w.PotIndex]
		if len(w.PlayerIDs) == 0 {
			continue
		}
		share := pot.Amount / len(w.PlayerIDs)
		remainder := pot.Amount % len(w.PlayerIDs)

		for i, pid := range w.PlayerIDs {
			player := room.FindPlayer(pid)
			if player != nil {
				winAmount := share
				if i == 0 {
					winAmount += remainder
				}
				player.Stack += winAmount
			}
		}
	}

	room.Round.IsComplete = true
	room.Status = entity.RoomStatusWaiting
	room.Round = nil

	if room.Config.GameMode == entity.GameModeTournament {
		for i := range room.Players {
			if room.Players[i].Stack <= 0 && room.Players[i].Status != entity.PlayerStatusEliminated {
				room.Players[i].Status = entity.PlayerStatusEliminated
			}
		}
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) PauseGame(ctx context.Context, roomID, playerID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.HostPlayerID != playerID {
		return nil, entity.ErrNotHost
	}

	if room.Status == entity.RoomStatusPlaying {
		room.Status = entity.RoomStatusPaused
	} else if room.Status == entity.RoomStatusPaused {
		room.Status = entity.RoomStatusPlaying
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) Rebuy(ctx context.Context, roomID, playerID string, amount int) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Config.GameMode != entity.GameModeCash {
		return nil, entity.ErrInvalidAction
	}

	player := room.FindPlayer(playerID)
	if player == nil {
		return nil, entity.ErrPlayerNotFound
	}

	if amount <= 0 {
		amount = room.Config.StartingStack
	}

	player.Stack += amount
	if player.Status == entity.PlayerStatusEliminated {
		player.Status = entity.PlayerStatusWaiting
	}

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) KickPlayer(ctx context.Context, roomID, hostID, targetID string) (*entity.Room, error) {
	room, err := uc.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.HostPlayerID != hostID {
		return nil, entity.ErrNotHost
	}

	newPlayers := make([]entity.Player, 0, len(room.Players))
	for _, p := range room.Players {
		if p.ID != targetID {
			newPlayers = append(newPlayers, p)
		}
	}
	room.Players = newPlayers

	if err := uc.repo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *GameUseCase) nextDealer(room *entity.Room, activePlayers []entity.Player) int {
	if room.Round == nil || room.RoundCount == 1 {
		return activePlayers[0].Seat
	}
	prevDealer := room.Round.DealerSeat
	for _, p := range activePlayers {
		if p.Seat > prevDealer {
			return p.Seat
		}
	}
	return activePlayers[0].Seat
}

func (uc *GameUseCase) postBlinds(room *entity.Room, round *entity.Round, activePlayers []entity.Player) {
	if len(activePlayers) < 2 {
		return
	}

	dealerIdx := -1
	for i, p := range activePlayers {
		if p.Seat == round.DealerSeat {
			dealerIdx = i
			break
		}
	}

	var sbIdx, bbIdx int
	if len(activePlayers) == 2 {
		sbIdx = dealerIdx
		bbIdx = (dealerIdx + 1) % len(activePlayers)
	} else {
		sbIdx = (dealerIdx + 1) % len(activePlayers)
		bbIdx = (dealerIdx + 2) % len(activePlayers)
	}

	sbPlayer := room.FindPlayer(activePlayers[sbIdx].ID)
	sbAmount := min(round.SmallBlind, sbPlayer.Stack)
	sbPlayer.Stack -= sbAmount
	for i := range round.PlayerStates {
		if round.PlayerStates[i].PlayerID == sbPlayer.ID {
			round.PlayerStates[i].Bet = sbAmount
			round.PlayerStates[i].TotalBet = sbAmount
			if sbPlayer.Stack == 0 {
				round.PlayerStates[i].AllIn = true
			}
			break
		}
	}

	bbPlayer := room.FindPlayer(activePlayers[bbIdx].ID)
	bbAmount := min(round.BigBlind, bbPlayer.Stack)
	bbPlayer.Stack -= bbAmount
	for i := range round.PlayerStates {
		if round.PlayerStates[i].PlayerID == bbPlayer.ID {
			round.PlayerStates[i].Bet = bbAmount
			round.PlayerStates[i].TotalBet = bbAmount
			if bbPlayer.Stack == 0 {
				round.PlayerStates[i].AllIn = true
			}
			break
		}
	}

	round.CurrentBet = bbAmount
	round.MinRaise = bbAmount
}

func (uc *GameUseCase) setFirstToAct(round *entity.Round, activePlayers []entity.Player) {
	dealerIdx := -1
	for i, p := range activePlayers {
		if p.Seat == round.DealerSeat {
			dealerIdx = i
			break
		}
	}

	var firstIdx int
	if len(activePlayers) == 2 {
		firstIdx = dealerIdx
	} else {
		firstIdx = (dealerIdx + 3) % len(activePlayers)
	}

	for i := 0; i < len(activePlayers); i++ {
		idx := (firstIdx + i) % len(activePlayers)
		ps := findPlayerState(round, activePlayers[idx].ID)
		if ps != nil && !ps.AllIn && !ps.Folded {
			round.CurrentTurn = activePlayers[idx].ID
			return
		}
	}
}

func setFirstToActPostflop(round *entity.Round, activePlayers []entity.Player) {
	dealerIdx := -1
	for i, p := range activePlayers {
		if p.Seat == round.DealerSeat {
			dealerIdx = i
			break
		}
	}

	for i := 1; i <= len(activePlayers); i++ {
		idx := (dealerIdx + i) % len(activePlayers)
		ps := findPlayerState(round, activePlayers[idx].ID)
		if ps != nil && !ps.AllIn && !ps.Folded {
			round.CurrentTurn = activePlayers[idx].ID
			return
		}
	}
}

func nextStreet(current entity.Street) entity.Street {
	switch current {
	case entity.StreetPreflop:
		return entity.StreetFlop
	case entity.StreetFlop:
		return entity.StreetTurn
	case entity.StreetTurn:
		return entity.StreetRiver
	case entity.StreetRiver:
		return entity.StreetShowdown
	default:
		return ""
	}
}

func collectBets(round *entity.Round) {
	for i := range round.PlayerStates {
		round.PlayerStates[i].Bet = 0
	}
}

func CalculatePots(round *entity.Round) []entity.Pot {
	type playerBet struct {
		id       string
		totalBet int
		folded   bool
	}

	var bets []playerBet
	for _, ps := range round.PlayerStates {
		bets = append(bets, playerBet{
			id:       ps.PlayerID,
			totalBet: ps.TotalBet,
			folded:   ps.Folded,
		})
	}

	sort.Slice(bets, func(i, j int) bool {
		return bets[i].totalBet < bets[j].totalBet
	})

	var pots []entity.Pot
	prevLevel := 0

	for _, b := range bets {
		if b.totalBet <= prevLevel {
			continue
		}

		level := b.totalBet
		potAmount := 0
		var eligible []string

		for _, pb := range bets {
			contribution := min(pb.totalBet, level) - min(pb.totalBet, prevLevel)
			if contribution < 0 {
				contribution = 0
			}
			potAmount += contribution
			if !pb.folded && pb.totalBet >= level {
				eligible = append(eligible, pb.id)
			}
		}

		for _, pb := range bets {
			if !pb.folded && pb.totalBet == level {
				found := false
				for _, e := range eligible {
					if e == pb.id {
						found = true
						break
					}
				}
				if !found {
					eligible = append(eligible, pb.id)
				}
			}
		}

		if potAmount > 0 {
			pots = append(pots, entity.Pot{
				Amount:      potAmount,
				EligibleIDs: eligible,
			})
		}

		prevLevel = level
	}

	if len(pots) == 0 {
		pots = append(pots, entity.Pot{Amount: 0})
	}

	return pots
}

func findPlayerState(round *entity.Round, playerID string) *entity.PlayerState {
	for i := range round.PlayerStates {
		if round.PlayerStates[i].PlayerID == playerID {
			return &round.PlayerStates[i]
		}
	}
	return nil
}
