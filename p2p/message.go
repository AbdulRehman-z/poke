package p2p

type Message struct {
	From    string
	Payload any
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From:    from,
		Payload: payload,
	}
}

type MessagePeerList struct {
	Peers []string
}
