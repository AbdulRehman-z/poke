package p2p

import (
	"bytes"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
)

type TCPPeer struct {
	conn     *net.TCPConn
	outbound bool
}

func NewTCPPeer(conn *net.TCPConn) *TCPPeer {
	return &TCPPeer{
		conn: conn,
	}
}

func (p *TCPPeer) Send(msg []byte) error {
	_, err := p.conn.Write(msg)
	return err
}

type Message struct {
	From    net.Addr
	To      net.Addr
	Payload io.Reader
}

type TCPTransportOpts struct {
	Laddr     *net.TCPAddr
	OnPeer    OnPeer
	Handshake HandshakeFunc
}

type TCPTransport struct {
	*TCPTransportOpts
	tcpListener *net.TCPListener
	Handler     Handler

	// mu          sync.RWMutex
	addPeerChan chan *TCPPeer
	delPeerCh   chan *TCPPeer
	msgChan     chan *Message
}

// // OnPeer implements Transport.
// func (t *TCPTransport) OnPeer(*TCPPeer) error {
// 	panic("unimplemented")
// }

func NewTCPTransport(opts *TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		Handler:          &DefaultHandler{},
		addPeerChan:      make(chan *TCPPeer),
		delPeerCh:        make(chan *TCPPeer),
		msgChan:          make(chan *Message),
	}
}

// Addr implements the Transport interface, it returns the address of the transport
func (t *TCPTransport) Addr() string {
	return t.Laddr.AddrPort().String()
}

// Consume implements the Transport interface, it returns the message channel
func (t *TCPTransport) Consume() <-chan *Message {
	return t.msgChan
}

// AddPeer implements the Transport interface, it returns the peer channel
func (t *TCPTransport) AddPeer() <-chan *TCPPeer {
	return t.addPeerChan
}

// DelPeer implements the Transport interface, it returns the peer channel
func (t *TCPTransport) DelPeer() <-chan *TCPPeer {
	return t.delPeerCh
}

// ListenAndAccept implements the Transport interface, it listens for incoming connections
func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.ListenTCP("tcp", t.Laddr)
	if err != nil {
		slog.Error("ERR tcp listen", "err", err)
	}
	t.tcpListener = ln
	slog.Info("tcp listening", "addr", t.Addr())

	go t.tcpAcceptLoop()
	return nil
}

// Dial implements the Transport interface, it dials a connection to the given port and handles
// the connection that is returned by the dial i.e it reads from the connection
func (t *TCPTransport) Dial(port int, gv GameVariant, version string) error {
	conn, err := net.DialTCP("tcp", nil,
		&net.TCPAddr{
			IP:   net.ParseIP("localhost"),
			Port: port,
		})
	if err != nil {
		slog.Error("ERR dial failed", "port", port)
	}
	peer := &TCPPeer{
		conn:     conn,
		outbound: true,
	}
	t.addPeerChan <- peer

	return SendHandshake(peer, gv, version)
}

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
		}

		peer := &TCPPeer{
			conn:     conn,
			outbound: false,
		}
		t.addPeerChan <- peer
	}
}

// handleConn reads from the connection and sends the data to the message channel
// Note: this is executed for every new connection in separate go routine handling specifically
// that connection i.e. it ONLY reads from that connection
func (t *TCPTransport) HandlePeer(peer *TCPPeer, variant GameVariant, version string) error {
	for {
		buf := make([]byte, 1024)
		n, err := peer.conn.Read(buf)
		if errors.Is(err, net.ErrClosed) {
			return err
		}
		if err != nil {
			peer.conn.Close()
			t.delPeerCh <- peer
			return err
		}

		log.Println("Read Loop Started")

		t.msgChan <- &Message{
			From:    peer.conn.RemoteAddr(),
			To:      peer.conn.LocalAddr(),
			Payload: bytes.NewReader(buf[:n]),
		}
	}
}
