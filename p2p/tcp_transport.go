package p2p

import (
	"bytes"
	"errors"
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

// Transport, every transport whether its udp,tcp or gRPC
// must implement this interface.
type Transport interface {
	ListenAndAccept() error
}

type TCPTransportOpts struct {
	Laddr *net.TCPAddr
}

type TCPTransport struct {
	TCPTransportOpts
	tcpListener *net.TCPListener
	Handler     Handler

	// mu          sync.RWMutex
	peers       map[net.Addr]*TCPPeer
	addPeerChan chan *TCPPeer
	msgChan     chan *Message
}

func NewTCPTransport(opts TCPTransportOpts) Transport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		Handler:          &DefaultHandler{},
		peers:            make(map[net.Addr]*TCPPeer),
		addPeerChan:      make(chan *TCPPeer),
		msgChan:          make(chan *Message),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.ListenTCP("tcp", t.Laddr)
	if err != nil {
		slog.Error("ERR tcp listen", "err", err)
	}
	t.tcpListener = ln

	go t.loop()
	go t.tcpAcceptLoop()

	return nil
}

func (t *TCPTransport) tcpAcceptLoop() {
	for {
		conn, err := t.tcpListener.AcceptTCP()
		log.Println("accepting")

		if errors.Is(err, net.ErrClosed) {
			slog.Error("conection didn't accepted", "err", err)
			return
		}
		if err != nil {
			slog.Error("ERR tcp accept", "err", err)
		}

		t.addPeerChan <- &TCPPeer{Conn: conn}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn *net.TCPConn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		log.Println("handling conn")

		if err != nil {
			slog.Error("ERR tcp read", "err", err)
		}

		t.msgChan <- &Message{
			From:    conn.RemoteAddr(),
			To:      conn.LocalAddr(),
			Payload: bytes.NewReader(buf[:n]),
		}
		slog.Info("tcp read", "msg", string(buf[:n]))
	}
}

func (t *TCPTransport) loop() {
	for {
		select {
		case peer := <-t.addPeerChan:
			t.peers[peer.RemoteAddr()] = peer
			peer.Send([]byte("You are accepted!"))
			slog.Info("peer added", "addr", peer.RemoteAddr())
		case msg := <-t.msgChan:
			if err := t.Handler.HandleMessage(msg); err != nil {
				slog.Error("ERR handle message", "err", err)
			}
		}
	}
}
