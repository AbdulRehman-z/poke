package p2p

import "net"

type Peer interface {
	net.Conn

	Send(b []byte) error
}

// Transport, every transport whether its udp,tcp or gRPC
// must implement this interface.
type Transport interface {
	ListenAndAccept() error
	Addr() string
	Dial(int) error
	Consume() <-chan *Message
	AddPeer() <-chan Peer
	DelPeer() <-chan Peer
}
