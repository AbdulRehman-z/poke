package p2p

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

func (g *Game) getNextGameStatus() GameStatus {
	switch GameStatus(g.currentStatus.Get()) {
	case GameStatusPreFlop:
		return GameStatusFlop
	case GameStatusFlop:
		return GameStatusTurn
	case GameStatusTurn:
		return GameStatusRiver
	default:
		panic("invalid status")
	}
}

type AtomicInt struct {
	Value int32
}

func NewAtomicInt(val int32) *AtomicInt {
	return &AtomicInt{
		Value: val,
	}
}
func (a *AtomicInt) String() string {
	return fmt.Sprintf("%d", a.Value)
}

func (a *AtomicInt) Get() int32 {
	return atomic.LoadInt32(&a.Value)
}

func (a *AtomicInt) Set(value int32) {
	atomic.StoreInt32(&a.Value, value)
}

func (a *AtomicInt) Inc() {
	currentValue := a.Get()
	a.Set(currentValue + 1)
}

type PlayerActionsRecv struct {
	mu          sync.RWMutex
	recvActions map[string]MessagePlayerAction
}

func NewPlayerActionsRecv() *PlayerActionsRecv {
	return &PlayerActionsRecv{
		recvActions: make(map[string]MessagePlayerAction),
	}
}

func (p *PlayerActionsRecv) addAction(from string, action MessagePlayerAction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.recvActions[from] = action
}

type PlayersReady struct {
	mu           sync.RWMutex
	recvStatutes map[string]bool
}

func NewPlayersReady() *PlayersReady {
	return &PlayersReady{
		recvStatutes: make(map[string]bool),
	}
}

func (pr *PlayersReady) addRecvStatus(from string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatutes[from] = true
}

func (pr *PlayersReady) len() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.recvStatutes)
}

func (pr *PlayersReady) clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatutes = make(map[string]bool)
}

type Game struct {
	listenAddr  string
	broadcastch chan BroadcastToPeers

	// currentStatus should be atomically accessable.
	currentStatus *AtomicInt
	// currentDealer should be atomically accessable.
	// NOTE: this will be -1 when the game is in a bootstrapped state.
	currentDealer       *AtomicInt
	currentPlayerAction *AtomicInt
	currentPlayerTurn   *AtomicInt

	playersReady       *PlayersReady
	playersActionsRecv *PlayerActionsRecv

	playersList PlayersList
}

func NewGame(addr string, bc chan BroadcastToPeers) *Game {
	g := &Game{
		listenAddr:          addr,
		currentStatus:       NewAtomicInt(0),
		playersReady:        NewPlayersReady(),
		broadcastch:         bc,
		playersList:         PlayersList{},
		currentDealer:       NewAtomicInt(0),
		currentPlayerTurn:   NewAtomicInt(0),
		currentPlayerAction: NewAtomicInt(0),
		playersActionsRecv:  NewPlayerActionsRecv(),
	}

	g.playersList = append(g.playersList, addr)

	go g.loop()

	return g
}

func (g *Game) SetSatutus(s GameStatus) {
	g.setStatus(s)
}

func (g *Game) setStatus(s GameStatus) {
	if s == GameStatusFlop {
		g.currentPlayerTurn.Inc()
	}

	// Only update the status when the status is different.
	if GameStatus(g.currentStatus.Get()) != s {
		g.currentStatus.Set(int32(s))
	}
}

func (g *Game) getCurrentDealerAddr() (string, bool) {
	currentDealerAddr := g.playersList[g.currentDealer.Get()]
	return currentDealerAddr, g.listenAddr == currentDealerAddr
}

func (g *Game) SetPlayerReady(from string) {
	logrus.WithFields(logrus.Fields{
		"we":     g.listenAddr,
		"player": from,
	}).Info("setting player status to ready")

	g.playersReady.addRecvStatus(from)

	// If we don't have enough players the round cannot be started.
	if g.playersReady.len() < 2 {
		return
	}

	// In the case we have enough players. hence, the round can be started.
	// FIXME:(@AbdulRehman-z)
	// g.playersReady.clear()

	// we need to check if we are the dealer of the current round.
	if _, ok := g.getCurrentDealerAddr(); ok {
		g.InitiateShuffleAndDeal()
	}
}

func (g *Game) canTakeAction(from string) bool {
	currentPlayerAddr := g.playersList[g.currentPlayerTurn.Get()]
	return currentPlayerAddr == from
}

func (g *Game) handlePlayerAction(from string, msg MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player %s  taking his action before his turn", from)
	}
	if GameStatus(g.currentStatus.Get()) != msg.CurrentGameStatus {
		return fmt.Errorf("player status mismatched got = %d : expected = %d || we = %s, from = %s", msg.CurrentGameStatus, GameStatus(g.currentStatus.Get()), g.listenAddr, from)
	}
	logrus.WithFields(logrus.Fields{
		"we":   g.listenAddr,
		"from": from,
	}).Info("recv player action")

	//Every player in this case should need to set the current game status to next one!
	if g.playersList[g.currentDealer.Get()] == from {
		g.currentStatus.Set(int32(g.getNextGameStatus()))
	}

	g.incNextPlayer()

	g.playersActionsRecv.addAction(from, msg)
	return nil
}

func (g *Game) TakeAction(action PlayerAction, value int) (err error) {
	if !g.canTakeAction(g.listenAddr) {
		return fmt.Errorf("player %s taking his action before his turn", g.listenAddr)
	}

	g.currentPlayerAction.Set(int32(action))

	// if we are the dealer that just took an action, than we can just go to next round.
	if g.listenAddr == g.playersList[g.currentDealer.Get()] {
		g.currentStatus.Set(int32(g.getNextGameStatus()))
	}

	a := MessagePlayerAction{
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Action:            action,
		Value:             value,
	}

	g.sendToPlayers(a, g.getOtherPlayers()...)

	g.incNextPlayer()
	return
}

func (g *Game) bet(value int) error {
	action := MessagePlayerAction{
		Action: PlayerActionBet,
		Value:  value,
	}

	g.sendToPlayers(action, g.getOtherPlayers()...)
	return nil
}

func (g *Game) check() error {
	g.SetSatutus(GameStatusChecked)

	action := MessagePlayerAction{
		CurrentGameStatus: GameStatusChecked,
		Action:            PLayerActionCheck,
	}

	g.sendToPlayers(action, g.getOtherPlayers()...)

	return nil
}

func (g *Game) fold() error {
	g.setStatus(GameStatusFolded)

	action := MessagePlayerAction{
		Action:            PlayerActionFold,
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
	}

	g.sendToPlayers(action, g.getOtherPlayers()...)
	return nil
}

func (g *Game) incNextPlayer() {
	if len(g.playersList)-1 == int(g.currentPlayerTurn.Get()) {
		g.currentPlayerTurn.Set(0)
		return
	}

	g.currentPlayerTurn.Inc()
}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {
	prevPlayerAddr := g.playersList[g.getPrevPositionOnTable()]
	if from != prevPlayerAddr {
		return fmt.Errorf("received encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayerAddr)
	}

	_, isDealer := g.getCurrentDealerAddr()
	if isDealer && from == prevPlayerAddr {
		g.setStatus(GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		logrus.Info("shuffle round complete")
		return nil

	}
	dealToNextPlayer := g.playersList[g.getNextPositionOnTable()]

	logrus.WithFields(logrus.Fields{
		"recvFromPlayer":  from,
		"we":              g.listenAddr,
		"dealingToPlayer": dealToNextPlayer,
	}).Info("received cards and going to shuffle")

	// TODO:(@AbdulRehman-z) encryption and shuffle
	// TODO: get this player out of a deterministic (sorted) list.

	g.sendToPlayers(MessageEncCards{Deck: [][]byte{}}, dealToNextPlayer)
	g.setStatus(GameStatusDealing)
	g.currentPlayerTurn.Inc()

	return nil
}

func (g *Game) InitiateShuffleAndDeal() {
	dealToPlayerAddr := g.playersList[g.getNextPositionOnTable()]
	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncCards{Deck: [][]byte{}}, dealToPlayerAddr)
	g.currentPlayerTurn.Inc()

	logrus.WithFields(logrus.Fields{
		"we": g.listenAddr,
		"to": dealToPlayerAddr,
	}).Info("dealing cards")

}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)
	g.setStatus(GameStatusReady)
}

func (g *Game) sendToPlayers(payload any, addr ...string) {
	g.broadcastch <- BroadcastToPeers{
		To:      addr,
		Payload: payload,
	}

	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"player":  addr,
		"we":      g.listenAddr,
	}).Info("sending payload to player")
}

func (g *Game) AddPlayer(from string) {
	// If the player is being added to the game. We are going to assume
	// that he is ready to play.
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)
}

func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		<-ticker.C

		currentDealerAddr, _ := g.getCurrentDealerAddr()
		logrus.WithFields(logrus.Fields{
			"we":                  g.listenAddr,
			"players":             g.playersList,
			"status":              GameStatus(g.currentStatus.Get()),
			"currentDealer":       currentDealerAddr,
			"nextPlayerTurn":      g.currentPlayerTurn,
			"currentPlayerAction": PlayerAction(g.currentPlayerAction.Get()),
			"playerActionsRccv":   g.playersActionsRecv.recvActions,
		}).Info()
	}
}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList {
		if addr == g.listenAddr {
			continue
		}
		players = append(players, addr)
	}

	return players
}

// getPositionOnTable return the index of our own position on the table.
func (g *Game) getPositionOnTable() int {
	for i := 0; i < len(g.playersList); i++ {
		if g.playersList[i] == g.listenAddr {
			return i
		}
	}

	panic("player does not exist in the playersList; that should not happen!!!")
}

func (g *Game) getPrevPositionOnTable() int {
	ourPosition := g.getPositionOnTable()

	// if we are the in the first position on the table we need to return the last
	// index of the PlayersList.
	if ourPosition == 0 {
		return len(g.playersList) - 1
	}

	return ourPosition - 1
}

// getNextPositionOnTable returns the index of the next player in the PlayersList.
func (g *Game) getNextPositionOnTable() int {
	ourPosition := g.getPositionOnTable()

	// check if we are on the last position of the table, if so return first index 0.
	if ourPosition == len(g.playersList)-1 {
		return 0
	}

	return ourPosition + 1
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
