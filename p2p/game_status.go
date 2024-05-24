package p2p

// enum for GameStatus
type GameStatus int32

const (
	GameStatusConnected GameStatus = iota
	GameStatusReady
	GameStatusDealing
	GameStatusFolded
	GameStatusChecked
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "CONNECTED"
	case GameStatusReady:
		return "PLAYER READY"
	case GameStatusDealing:
		return "DEALING"
	case GameStatusFolded:
		return "FOLDED"
	case GameStatusChecked:
		return "CHECKED"
	case GameStatusPreFlop:
		return "PRE FLP"
	case GameStatusFlop:
		return "FLOP"
	case GameStatusTurn:
		return "TURN"
	case GameStatusRiver:
		return "RIVER"
	default:
		return "unknown"
	}
}

// enum for PlayerAction
type PlayerAction byte

const (
	PlayerActionFold PlayerAction = iota + 1 // 1
	PLayerActionCheck
	PlayerActionBet
)

func (a PlayerAction) String() string {
	switch a {
	case PLayerActionCheck:
		return "CHECKED"
	case PlayerActionFold:
		return "FOLD"
	case PlayerActionBet:
		return "BET"
	default:
		return "invalid player action"
	}
}
