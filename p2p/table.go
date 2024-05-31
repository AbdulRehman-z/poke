package p2p

import (
	"fmt"
	"sync"
)

type Player struct {
	addr          string
	currentAction PlayerAction
	gameStatus    GameStatus
	tablePosition int
}

func NewPLayer(addr string) *Player {
	return &Player{
		addr:          addr,
		currentAction: PlayerActionNone,
		gameStatus:    GameStatusConnected,
		tablePosition: -1,
	}
}

type Table struct {
	lock     sync.RWMutex
	seats    map[int]*Player
	maxSeats int
}

func NewTable(seats int) *Table {
	return &Table{
		seats:    make(map[int]*Player, seats),
		maxSeats: seats,
	}
}

func (t *Table) AddPlayer(addr string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if len(t.seats) == t.maxSeats {
		return fmt.Errorf("table is full")
	}

	position := t.getNextAvaliableSeat()
	player := NewPLayer(addr)
	player.tablePosition = position

	t.seats[position] = player

	return nil
}

func (t *Table) GetPlayerByAddr(addr string) (*Player, error) {
	for k := range t.seats {
		player, ok := t.seats[k]
		if player.addr == addr {
			if ok {
				return player, nil
			}
		}
	}

	return nil, fmt.Errorf("player can't be found %s", addr)
}

func (t *Table) RemovePlayerByAddr(addr string) error {
	for k := range t.seats {
		player, ok := t.seats[k]
		if ok {
			if player.addr == addr {
				delete(t.seats, k)
				return nil
			}
		}
	}

	return fmt.Errorf("player can't be deleted %s", addr)
}

func (t *Table) GetPlayerBeforeMe(addr string) (*Player, error) {
	currentPlayer, err := t.GetPlayerByAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("can't get the player %s", addr)
	}

	i := currentPlayer.tablePosition - 1
	if i <= 0 {
		i = t.maxSeats
	}
	for {
		nextPlayer, ok := t.seats[i]
		if ok {
			if nextPlayer.addr == currentPlayer.addr {
				return nil, fmt.Errorf("%s is the only player in the table", addr)
			}

			return nextPlayer, nil
		}
		i--

	}
}

func (t *Table) GetPlayerAfterMe(addr string) (*Player, error) {
	currentPlayer, err := t.GetPlayerByAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("can't get the player %s", addr)
	}

	i := currentPlayer.tablePosition + 1
	if t.maxSeats == i {
		i = 0
	}
	for {
		nextPlayer, ok := t.seats[i]
		if ok {
			if nextPlayer.addr == currentPlayer.addr {
				return nil, fmt.Errorf("%s is the only player in the table", addr)
			}

			return nextPlayer, nil
		}
		i++
	}
}

func (t *Table) getNextAvaliableSeat() int {
	for k := range t.maxSeats {
		if _, ok := t.seats[k]; !ok {
			return k
		}
	}
	panic("no free seat avaliable")
}
