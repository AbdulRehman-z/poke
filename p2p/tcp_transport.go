package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
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

func (t *TCPTransport) Addr() string {
	return t.Laddr.AddrPort().String()
}

func (t *TCPTransport) Consume() <-chan *Message {
	return t.msgChan
}

func (t *TCPTransport) AddPeer() <-chan Peer {
	return t.addPeerChan
}

func (t *TCPTransport) DelPeer() <-chan Peer {
	return t.delPeerCh
}

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

func (t *TCPTransport) Dial(port int) error {
	conn, err := net.DialTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 9000,
	},
		&net.TCPAddr{
			IP:   net.ParseIP("localhost"),
			Port: port,
		})
	if err != nil {
		slog.Error("ERR dial failed", "port", port)
	}
	peer := &TCPPeer{
		Conn: conn,
	}

	fmt.Println(peer.Conn.RemoteAddr().String())
	t.addPeerChan <- peer
	// go t.handleConn(conn)
	return nil
}

func (t *TCPTransport) tcpAcceptLoop() {
	for {
		conn, err := t.tcpListener.AcceptTCP()
		log.Printf("%s accepted %s\n", t.Laddr.String(), conn.RemoteAddr().String())
		if errors.Is(err, net.ErrClosed) {
			slog.Error("conection didn't accepted", "err", err)
			return
		}
		if err != nil {
			slog.Error("ERR tcp accept", "err", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn *net.TCPConn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
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
