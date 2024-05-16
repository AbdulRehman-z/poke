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
