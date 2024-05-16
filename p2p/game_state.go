package p2p

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "CONNECTED"
	case GameStatusReady:
		return "PLAYER READY"
	case GameStatusReceivingCards:
		return "RECEIVING CARDS"
	case GameStatusShuffleEncryptANdDeal:
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
	GameStatusConnected GameStatus = iota
	GameStatusReady
	GameStatusReceivingCards
	GameStatusShuffleEncryptANdDeal
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type Game struct {
	listenAddr    string
	broadcastCh   chan BroadcastToPeers
	currentStatus GameStatus
	playersList   PlayersList
	playersReady  *PlayersReady
}

type PlayersReady struct {
	recvStatutesLock sync.RWMutex
	recvStatutes     map[string]bool
}

func NewGame(addr string, broadcastCh chan BroadcastToPeers) *Game {
	g := &Game{
		listenAddr:    addr,
		broadcastCh:   broadcastCh,
		currentStatus: GameStatusConnected,
		playersList:   PlayersList{},
		playersReady:  NewPlayersReady(),
	}

	go g.loop()
	return g
}

func NewPlayersReady() *PlayersReady {
	return &PlayersReady{
		recvStatutes: make(map[string]bool),
	}
}

func (g *Game) setStatus(s GameStatus) {
	if g.currentStatus != s {
		atomic.AddInt32((*int32)(&g.currentStatus), int32(s))
	}
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.SendToPlayers(MessageReady{}, g.getOtherPlayers()...)
	g.setStatus(GameStatusReady)

}

func (g *Game) SendToPlayers(payload any, addr ...string) {
	g.broadcastCh <- BroadcastToPeers{
		To:      addr,
		Payload: payload,
	}
}

func (p *PlayersReady) len() int {
	p.recvStatutesLock.Lock()
	defer p.recvStatutesLock.Unlock()

	return len(p.recvStatutes)
}

func (p *PlayersReady) addRecvStatus(from string) {
	p.recvStatutesLock.Lock()
	defer p.recvStatutesLock.Unlock()

	p.recvStatutes[from] = true
}

func (p *PlayersReady) clear() {
	p.recvStatutesLock.Lock()
	defer p.recvStatutesLock.Unlock()

	p.recvStatutes = make(map[string]bool)
}

func (g *Game) AddNewPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)
	g.playersReady.addRecvStatus(from)
}

type PlayersList []string

func (list PlayersList) Len() int { return len(list) }
func (list PlayersList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list PlayersList) Less(i, j int) bool {
	portI, _ := strconv.Atoi(list[i][1:])
	portJ, _ := strconv.Atoi(list[j][1:])

	return portI < portJ
}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList {
		if addr != g.listenAddr {
			players = append(players, addr)
		}
	}

	return players
}

func (g *Game) loop() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		<-ticker.C
		logrus.WithFields(logrus.Fields{
			"me":          g.listenAddr,
			"status":      g.currentStatus,
			"playerslist": g.playersList,
		}).Info("Game state")
	}
}

// type Player struct {
// 	Status GameStatus
// }

// func (p *Player) String() string {
// 	return fmt.Sprintf("%s:%s", p.ListenAddr, p.Status)
// }

// type GameState struct {
// 	listenAddr  string
// 	broadcastch chan BroadcastToPeers
// 	isDealer    bool // should be atomic accessable !

// 	gameStatus GameStatus // should be atomic accessable !

// 	playersList PlayersList
// 	playersLock sync.RWMutex
// 	players     map[string]*Player
// }

// func NewGameState(addr string, broadcastch chan BroadcastToPeers) *GameState {
// 	g := &GameState{
// 		listenAddr:  addr,
// 		broadcastch: broadcastch,
// 		isDealer:    false,
// 		gameStatus:  GameStatusWaitingForCards,
// 		players:     make(map[string]*Player),
// 	}

// 	g.AddNewPlayer(addr, GameStatusWaitingForCards)

// 	go g.loop()

// 	return g
// }

// // TODO:(@AbdulREhman-z) Check other read and write occurences of the GameStatus!
// func (g *GameState) SetStatus(s GameStatus) {
// 	// Only update the status when the status is different.
// 	if g.gameStatus != s {
// 		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(s))
// 		g.SetPlayerStatus(g.listenAddr, s)
// 	}
// }

// func (g *GameState) playersWaitingForCards() int {
// 	totalPlayers := 0
// 	for i := 0; i < len(g.playersList); i++ {
// 		if g.playersList[i].Status == GameStatusWaitingForCards {
// 			totalPlayers++
// 		}
// 	}
// 	return totalPlayers
// }

// func (g *GameState) CheckNeedDealCards() {
// 	playersWaiting := g.playersWaitingForCards()

// 	if playersWaiting == len(g.players) &&
// 		g.isDealer &&
// 		g.gameStatus == GameStatusWaitingForCards {

// 		logrus.WithFields(logrus.Fields{
// 			"addr": g.listenAddr,
// 		}).Info("need to deal cards")

// 		g.InitiateShuffleAndDeal()
// 	}
// }

// func (g *GameState) GetPlayersWithStatus(s GameStatus) []string {
// 	players := []string{}
// 	for addr, player := range g.players {
// 		if player.Status == s {
// 			players = append(players, addr)
// 		}
// 	}
// 	return players
// }

// // getPositionOnTable return the index of our own position on the table.
// func (g *GameState) getPositionOnTable() int {
// 	for i := 0; i < len(g.playersList); i++ {
// 		if g.playersList[i].ListenAddr == g.listenAddr {
// 			return i
// 		}
// 	}

// 	panic("player does not exist in the playersList; that should not happen!!!")
// }

// func (g *GameState) getPrevPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()

// 	// if we are the in the first position on the table we need to return the last
// 	// index of the PlayersList.
// 	if ourPosition == 0 {
// 		return len(g.playersList) - 1
// 	}

// 	return ourPosition - 1
// }

// // getNextPositionOnTable returns the index of the next player in the PlayersList.
// func (g *GameState) getNextPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()

// 	// check if we are on the last position of the table, if so return first index 0.
// 	if ourPosition == len(g.playersList)-1 {
// 		return 0
// 	}

// 	return ourPosition + 1
// }

// func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
// 	g.SetPlayerStatus(from, GameStatusShuffleEncryptANdDeal)

// 	prevPlayer := g.playersList[g.getPrevPositionOnTable()]
// 	if g.isDealer && from == prevPlayer.ListenAddr {
// 		logrus.Info("shuffle round complete")
// 		return nil
// 	}

// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

// 	logrus.WithFields(logrus.Fields{
// 		"recvFromPlayer":  from,
// 		"we":              g.listenAddr,
// 		"dealingToPlayer": dealToPlayer,
// 	}).Info("received cards and going to shuffle")

// 	// TODO:(@AbdulREhman-z) encryption and shuffle
// 	// TODO: get this player out of a deterministic (sorted) list.

// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncCards{Deck: [][]byte{}})
// 	g.SetStatus(GameStatusShuffleEncryptANdDeal)

// 	fmt.Printf("%s => setting my own status to %s\n", g.listenAddr, GameStatusShuffleEncryptANdDeal)

// 	return nil
// }

// // InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
// func (g *GameState) InitiateShuffleAndDeal() {
// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

// 	g.SetStatus(GameStatusShuffleEncryptANdDeal)
// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncCards{Deck: [][]byte{}})
// }

// func (g *GameState) SendToPlayer(addr string, payload any) {
// 	g.broadcastch <- BroadcastToPeers{
// 		To:      []string{addr},
// 		Payload: payload,
// 	}

// 	logrus.WithFields(logrus.Fields{
// 		"payload": payload,
// 		"player":  addr,
// 	}).Info("sending payload to player")
// }

// func (g *GameState) SendToPlayersWithStatus(payload any, s GameStatus) {
// 	players := g.GetPlayersWithStatus(s)

// 	g.broadcastch <- BroadcastToPeers{
// 		To:      players,
// 		Payload: payload,
// 	}

// 	logrus.WithFields(logrus.Fields{
// 		"payload": payload,
// 		"players": players,
// 	}).Info("sending to players")
// }

// func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
// 	player, ok := g.players[addr]
// 	if !ok {
// 		panic("player could not be found, altough it should exist")
// 	}

// 	player.Status = status

// 	g.CheckNeedDealCards()
// }

// func (g *GameState) AddNewPlayer(addr string, status GameStatus) {
// 	g.playersLock.Lock()
// 	defer g.playersLock.Unlock()

// 	player := &Player{
// 		ListenAddr: addr,
// 	}
// 	g.players[addr] = player
// 	g.playersList = append(g.playersList, player)
// 	sort.Sort(g.playersList)

// 	// Set the player status also when we add the player!
// 	g.SetPlayerStatus(addr, status)

// 	logrus.WithFields(logrus.Fields{
// 		"addr":   addr,
// 		"status": status,
// 	}).Info("new player joined")
// }

// func (g *GameState) loop() {
// 	ticker := time.NewTicker(time.Second * 5)

// 	for {
// 		<-ticker.C
// 		logrus.WithFields(logrus.Fields{
// 			"we":      g.listenAddr,
// 			"players": g.playersList,
// 			"status":  g.gameStatus,
// 		}).Info()
// 	}
// }
