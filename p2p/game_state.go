package p2p

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type GameStatus uint32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaitingForCards:
		return "WAITING FOR CARDS"
	case GameStatusDealing:
		return "DEALING"
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

const (
	GameStatusWaitingForCards GameStatus = iota
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type GameState struct {
	isDealer   bool       // should be atomic accessable !
	gameStatus GameStatus // should be atomic accessable !

	playersLock sync.RWMutex
	players     map[string]*Player
}

type Player struct {
	Status GameStatus
}

func (g *GameState) AddNewPlayer(addr string, gameStatus GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	g.players[addr] = &Player{
		Status: gameStatus,
	}

	logrus.WithFields(logrus.Fields{
		"player joined": addr,
		"player status": gameStatus,
	}).Info("new player joined")
}

func (g *GameState) LenPlayersConnected() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()

	return len(g.players)
}

func NewGameState() *GameState {
	g := &GameState{
		isDealer:   false,
		gameStatus: GameStatusWaitingForCards,
		players:    make(map[string]*Player),
	}

	go g.loop()
	return g
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ticker.C:
			logrus.WithFields(logrus.Fields{
				"players connected": g.LenPlayersConnected(),
				"game status":       g.gameStatus,
			}).Info()
		default:
		}
	}
}
