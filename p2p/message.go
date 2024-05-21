package p2p

import "github.com/AbdulRehman-z/poke/deck"

type Message struct {
	Payload any
	From    string
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From:    from,
		Payload: payload,
	}
}

type Handshake struct {
	Version     string
	GameVariant GameVariant
	GameStatus  GameStatus
	ListenAddr  string
}

type MessagePlayerAction struct {
	// CurrentGameStatus is the current game status of the player sending the message
	//status should be exactly same as ours
	CurrentGameStatus GameStatus
	// Action is the action the player is taking
	Action PlayerAction
	// Value is the value of the bet if any
	Value int
}

type MessagePeerList struct {
	Peers []string
}

type MessageCards struct {
	Deck deck.Deck
}

type MessageEncCards struct {
	Deck [][]byte
}

type BroadcastToPeers struct {
	To      []string
	Payload any
}

type MessageReady struct{}

func (msg MessageReady) String() string {
	return "READY"
}

type MessagePreFlop struct{}

func (msg MessagePreFlop) String() string {
	return "PREFLOP"
}
