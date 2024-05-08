package p2p

import "net"

type OnPeer func(*TCPPeer) error

type Peer interface {
	net.Conn

	Send(b []byte) error
}

// Transport, every transport whether its udp,tcp or gRPC
// must implement this interface.
type Transport interface {
	ListenAndAccept() error
	Addr() string
	Dial(int, GameVariant, string, string) error
	Consume() <-chan *Message
	AddPeer() chan *TCPPeer
	DelPeer() <-chan *TCPPeer
	HandlePeer(*TCPPeer, GameVariant, string) error
}
