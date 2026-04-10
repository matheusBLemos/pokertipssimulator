package event

type Type string

const (
	RoomStateChanged  Type = "room_state"
	PlayerLeft        Type = "player_left"
	RoundStarted      Type = "round_started"
	ActionPerformed   Type = "action_performed"
	StreetAdvanced    Type = "street_advanced"
	StackUpdated      Type = "stack_updated"
	Settlement        Type = "settlement"
	BlindLevelChanged Type = "blind_level_changed"
	ChipsTransferred  Type = "chips_transferred"
	GamePaused        Type = "game_paused"
	GameResumed       Type = "game_resumed"
)
