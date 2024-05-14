package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaitingForCards:
		return "WAITING FOR CARDS"
	case GameStatusReceivingCards:
		return "RECEIVING CARDS"
	case GameStatusShuffleAndEncrypt:
		return " SHUFFLING AND ENCRYPTING"
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
	GameStatusReceivingCards
	GameStatusShuffleAndEncrypt
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type Player struct {
	ListenAddr string
	Status     GameStatus
}

type GameState struct {
	ListenAddr string

	isDealer    bool       // should be atomic accessable !
	gameStatus  GameStatus // should be atomic accessable !
	playersList []*Player

	broadcastCh chan BroadcastToPeers

	playersWaitingForCards int32
	playersLock            sync.RWMutex
	players                map[string]*Player

	decksReceivedLock sync.RWMutex
	decksReceived     map[string]bool
}

func NewGameState(addr string, broadcastCh chan BroadcastToPeers) *GameState {
	g := &GameState{
		ListenAddr:    addr,
		isDealer:      false,
		gameStatus:    GameStatusWaitingForCards,
		playersList:   make([]*Player, 0),
		players:       make(map[string]*Player),
		broadcastCh:   broadcastCh,
		decksReceived: map[string]bool{},
	}

	go g.loop()
	return g
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {

	dealToPlayer := g.playersList[0]

	g.SendToPlayer(dealToPlayer.ListenAddr, deck)
	g.SetStatus(GameStatusShuffleAndEncrypt)

	return nil
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) SetStatus(s GameStatus) {
	if g.gameStatus != s {
		atomic.StoreInt32((*int32)(&g.gameStatus), int32(s))
	}
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaitingForCards)

	if playersWaiting == int32(len(g.players)) && g.isDealer && g.gameStatus == GameStatusWaitingForCards {
		// panic("implement me")

		logrus.WithFields(logrus.Fields{
			"players waiting":   playersWaiting,
			"players connected": len(g.players),
		}).Info("deal cards")

		g.InitiateShuffleAndDeal()
	}
}

func (g *GameState) InitiateShuffleAndDeal() {
	dealToPlayer := g.playersList[0]

	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncCards{Deck: [][]byte{}})
	g.SetStatus(GameStatusShuffleAndEncrypt)
}

func (g *GameState) SendToPlayer(addr string, payload any) {
	g.broadcastCh <- BroadcastToPeers{
		To:      []string{addr},
		Payload: payload,
	}

	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"to":      addr,
		"from":    g.ListenAddr,
	}).Info("sending payload to player")
}

func (g *GameState) SendToPlayersWithStatus(msg MessageEncCards, status GameStatus) {
	players := g.GetPlayersWithStatus(status)

	g.broadcastCh <- BroadcastToPeers{
		To:      players,
		Payload: msg,
	}
}

func (g *GameState) GetPlayersWithStatus(s GameStatus) []string {
	players := []string{}
	for addr := range g.players {
		players = append(players, addr)
	}
	return players
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
	player, ok := g.players[addr]
	if !ok {
		// panic("player not found")
		return
	}
	player.Status = status
	g.CheckNeedDealCards()
}

func (g *GameState) AddNewPlayer(addr string, status GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	if status == GameStatusWaitingForCards {
		g.AddPlayerWaitingForCards()
	}
	player := &Player{
		ListenAddr: addr,
	}
	g.players[addr] = player
	g.playersList = append(g.playersList, player)

	g.SetPlayerStatus(addr, status)

	logrus.WithFields(logrus.Fields{
		"player joined": addr,
		"player status": status,
	}).Info("new player joined")
}

func (g *GameState) LenPlayersConnected() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()

	return len(g.players)
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ticker.C:
			logrus.WithFields(logrus.Fields{
				"me":                g.ListenAddr,
				"players connected": g.LenPlayersConnected(),
				"game status":       g.gameStatus,
			}).Info()
		default:
		}
	}
}
