package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayerInTable(t *testing.T) {
	maxSeats := 2
	table := NewTable(maxSeats)
	addr1 := ":2000"
	addr2 := ":3000"
	addr3 := ":4000"

	// add players
	err := table.AddPlayer(addr1)
	assert.NoError(t, err)
	err = table.AddPlayer(addr2)
	assert.NoError(t, err)
	err = table.AddPlayer(addr3) // table is already full at this point
	assert.NotNil(t, err)

	// get player
	player, err := table.GetPlayerByAddr(addr1)
	assert.NoError(t, err)
	assert.Equal(t, player.addr, addr1)

	// remove player
	err = table.RemovePlayerByAddr(addr1)
	assert.NoError(t, err)

	// get the same player again after removing it
	player, err = table.GetPlayerByAddr(addr1)
	assert.Nil(t, player)
	assert.Error(t, err)

	// get player with addr2
	player, err = table.GetPlayerByAddr(addr2)
	assert.Equal(t, player.addr, addr2)
	assert.Nil(t, err)

	// remove player with addr2
	err = table.RemovePlayerByAddr(addr2)
	assert.Nil(t, err)

	// table seats must be empty now
	assert.Len(t, table.seats, 0)
}

func TestGetPlayerBeforeMe(t *testing.T) {

	var (
		maxSeats = 4
		table    = NewTable(maxSeats)
		addr1    = ":2000"
		addr2    = ":3000"
		addr3    = ":4000"
		addr4    = ":5000"
	)

	// add players
	assert.Nil(t, table.AddPlayer(addr1))
	assert.Nil(t, table.AddPlayer(addr2))
	assert.Nil(t, table.AddPlayer(addr3))
	assert.Nil(t, table.AddPlayer(addr4))

	// get players before me
	player, _ := table.GetPlayerBeforeMe(addr1)
	assert.Equal(t, player.addr, addr4)
	player, _ = table.GetPlayerBeforeMe(addr4)
	assert.Equal(t, player.addr, addr3)

	// remove the player and then get player before me
	assert.Nil(t, table.RemovePlayerByAddr(addr1))
	player, _ = table.GetPlayerBeforeMe(addr2)
	assert.Equal(t, player.addr, addr4)
}

func TestGetPlayerAfterMe(t *testing.T) {
	var (
		maxSeats = 4
		table    = NewTable(maxSeats)
		addr1    = ":2000"
		addr2    = ":3000"
		addr3    = ":4000"
		addr4    = ":5000"
	)

	// add players
	assert.Nil(t, table.AddPlayer(addr1))
	assert.Nil(t, table.AddPlayer(addr2))
	assert.Nil(t, table.AddPlayer(addr3))
	assert.Nil(t, table.AddPlayer(addr4))

	// get next player
	player, _ := table.GetPlayerAfterMe(addr1)
	assert.Equal(t, player.addr, addr2)
	player, _ = table.GetPlayerAfterMe(addr2)
	assert.Equal(t, player.addr, addr3)
	player, _ = table.GetPlayerAfterMe(addr3)
	assert.Equal(t, player.addr, addr4)
	player, _ = table.GetPlayerAfterMe(addr4)
	assert.Equal(t, player.addr, addr1)

	// remove one player and the get next player from player map
	err := table.RemovePlayerByAddr(addr2)
	assert.NoError(t, err)
	player, err = table.GetPlayerAfterMe(addr1)
	assert.Nil(t, err)
	assert.Equal(t, player.addr, addr3)
}
