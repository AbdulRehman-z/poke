package p2p

func (gs GameStatus) String() string {
	switch gs {
	case GameStatusDealing:
		return "DEALING"
	case GameStatusWaiting:
		return "WAITING"
	case GameStatusPreFlop:
		return "PRE-FLOP"
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

type GameStatus uint32

const (
	GameStatusWaiting GameStatus = iota
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type GameState struct {
	isDealer   bool // should be atomic accessiable
	gameStatus GameStatus
}

func NewGameState() *GameState {
	return &GameState{}
}
