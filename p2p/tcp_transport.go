package p2p

import (
	"encoding/gob"
	"errors"
	"log/slog"
	"net"
)

type NetAddr string

func (n NetAddr) String() string  { return string(n) }
func (n NetAddr) Network() string { return "tcp" }

type TCPPeer struct {
	conn       *net.TCPConn
	outbound   bool
	listenAddr string
}

func NewTCPPeer(conn *net.TCPConn) *TCPPeer {
	return &TCPPeer{
		conn: conn,
	}
}

func (p *TCPPeer) Send(msg []byte) error {
	// log.Println(string(msg))
	_, err := p.conn.Write(msg)
	return err
}

type TCPTransport struct {
	Laddr       *net.TCPAddr
	tcpListener *net.TCPListener
	// Handler     Handler

	// mu          sync.RWMutex
	addPeerChan chan *TCPPeer
	delPeerCh   chan *TCPPeer
	msgChan     chan *Message
}

// // OnPeer implements Transport.
// func (t *TCPTransport) OnPeer(*TCPPeer) error {
// 	panic("unimplemented")
// }

func NewTCPTransport(addr *net.TCPAddr) *TCPTransport {
	return &TCPTransport{
		Laddr: addr,
		// Handler:          &DefaultHandler{},
		addPeerChan: make(chan *TCPPeer, 20),
		delPeerCh:   make(chan *TCPPeer),
		msgChan:     make(chan *Message),
	}
}

// Addr implements the Transport interface, it returns the address of the transport
func (t *TCPTransport) Addr() string {
	return t.Laddr.AddrPort().String()
}

// // Consume implements the Transport interface, it returns the message channel
// func (t *TCPTransport) Consume() <-chan *Message {
// 	return t.msgChan
// }

// // AddPeer implements the Transport interface, it returns the peer channel
// func (t *TCPTransport) AddPeer() chan *TCPPeer {
// 	return t.addPeerChan
// }

// // DelPeer implements the Transport interface, it returns the peer channel
// func (t *TCPTransport) DelPeer() <-chan *TCPPeer {
// 	return t.delPeerCh
// }

// ListenAndAccept implements the Transport interface, it listens for incoming connections
func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.ListenTCP("tcp", t.Laddr)
	if err != nil {
		slog.Error("ERR tcp listen", "err", err)
	}
	t.tcpListener = ln
	slog.Info("tcp listening", "addr", t.Laddr.String())

	go t.tcpAcceptLoop()
	return nil
}

// Dial implements the Transport interface, it dials a connection to the given port and handles
// the connection that is returned by the dial i.e it reads from the connection

// tcpAcceptLoop listens for incoming connections and adds them to the peer channel
func (t *TCPTransport) tcpAcceptLoop() {
	for {
		conn, err := t.tcpListener.AcceptTCP()
		if errors.Is(err, net.ErrClosed) {
			slog.Error("conection didn't accepted", "err", err)
			return
		}
		if err != nil {
			slog.Error("ERR tcp accept", "err", err)
			// return
		}

		peer := &TCPPeer{
			conn:     conn,
			outbound: false,
		}

		t.addPeerChan <- peer
		// fmt.Printf("%s added %s\n", peer.listenAddr, peer.conn.RemoteAddr())
	}
}

// handleConn reads from the connection and sends the data to the message channel
// Note: this is executed for every new connection in separate go routine handling specifically
// that connection i.e. it ONLY reads from that connection
func (t *TCPTransport) HandlePeer(peer *TCPPeer, variant GameVariant, version string) error {
	for {
		msg := new(Message)
		if err := gob.NewDecoder(peer.conn).Decode(msg); err != nil {
			return err
		}

		t.msgChan <- msg
	}
}
