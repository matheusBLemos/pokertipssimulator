package event

type Type string

const (
	RoomStateChanged    Type = "room_state"
	PlayerJoined        Type = "player_joined"
	PlayerLeft          Type = "player_left"
	PlayerReconnected   Type = "player_reconnected"
	RoundStarted        Type = "round_started"
	TurnChanged         Type = "turn_changed"
	ActionPerformed     Type = "action_performed"
	StreetAdvanced      Type = "street_advanced"
	PotsUpdated         Type = "pots_updated"
	StackUpdated        Type = "stack_updated"
	Settlement          Type = "settlement"
	RoundEnded          Type = "round_ended"
	BlindLevelChanged   Type = "blind_level_changed"
	ChipsTransferred    Type = "chips_transferred"
	GamePaused          Type = "game_paused"
	GameResumed         Type = "game_resumed"
)
