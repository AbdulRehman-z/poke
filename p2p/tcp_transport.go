package p2p

import (
	"bytes"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"time"
)

// type Peer interface {
// 	net.TCPConn
// }

type TCPPeer struct {
	net.Conn
}

func (p *TCPPeer) Send(msg []byte) error {
	_, err := p.Write(msg)
	return err
}

type Message struct {
	From    net.Addr
	To      net.Addr
	Payload io.Reader
}

type TCPTransportOpts struct {
	Laddr *net.TCPAddr
}

type TCPTransport struct {
	TCPTransportOpts
	tcpListener *net.TCPListener
	Handler     Handler

	// mu          sync.RWMutex
	addPeerChan chan Peer
	delPeerCh   chan Peer
	msgChan     chan *Message
}

func NewTCPTransport(opts TCPTransportOpts) Transport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		Handler:          &DefaultHandler{},
		addPeerChan:      make(chan Peer),
		delPeerCh:        make(chan Peer),
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
func (t *TCPTransport) AddPeer() <-chan Peer {
	return t.addPeerChan
}

// DelPeer implements the Transport interface, it returns the peer channel
func (t *TCPTransport) DelPeer() <-chan Peer {
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
func (t *TCPTransport) Dial(port int) error {
	conn, err := net.DialTCP("tcp", nil,
		&net.TCPAddr{
			IP:   net.ParseIP("localhost"),
			Port: port,
		})
	if err != nil {
		slog.Error("ERR dial failed", "port", port)
	}

	// we launch this go routine to handle any kind of arbitrary data that
	// the connection might send us
	go t.handleConn(conn)
	time.Sleep(1 * time.Second)
	conn.Write([]byte("hello from dialer"))
	return nil
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
			Conn: conn,
		}
		t.addPeerChan <- peer

		// we launch this go routine to handle any kind of arbitrary data that
		// we accept from the dialed connection or telnet connection
		go t.handleConn(conn)
	}
}

// handleConn reads from the connection and sends the data to the message channel
// Note: this is executed for every new connection in separate go routine handling specifically
// that connection i.e. it ONLY reads from that connection
func (t *TCPTransport) handleConn(conn *net.TCPConn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		log.Println("read")
		if err != nil {
			slog.Error("ERR tcp read", "err", err)
		}

		t.msgChan <- &Message{
			From:    conn.RemoteAddr(),
			To:      conn.LocalAddr(),
			Payload: bytes.NewReader(buf[:n]),
		}
	}
}
