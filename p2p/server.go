package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"sync"
)

type GameVariant uint8

func (gv GameVariant) String() string {
	switch gv {
	case TexasHoldings:
		return "TEXAS HOLDINGS"
	case Other:
		return "other"
	default:
		return "unknown"
	}
}

const (
	TexasHoldings GameVariant = iota
	Other
)

type ServerConfig struct {
	ListenAddr  *net.TCPAddr
	GameVariant GameVariant
	GameVersion string
	Transport   *TCPTransport
}

type Server struct {
	*ServerConfig

	mu      sync.RWMutex
	peers   map[net.Addr]*TCPPeer
	addPeer chan *TCPPeer
	delPeer chan *TCPPeer
	msgChan chan *Message
}

func NewServer(opts *ServerConfig) *Server {

	s := &Server{
		ServerConfig: opts,
		peers:        make(map[net.Addr]*TCPPeer),
		addPeer:      make(chan *TCPPeer, 20),
		delPeer:      make(chan *TCPPeer),
		msgChan:      make(chan *Message),
	}

	tcpTransport := NewTCPTransport(s.ListenAddr)
	s.Transport = tcpTransport
	tcpTransport.addPeerChan = s.addPeer
	tcpTransport.delPeerCh = s.delPeer
	tcpTransport.msgChan = s.msgChan
	return s
}

func (s *Server) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.loop()
	return nil
}

func (s *Server) Dial(port int, gv GameVariant, version string, serverLAddr string) error {
	// if t.tcpListener.Addr().String() == serverLAddr {
	// 	return nil
	// }

	// fmt.Printf("Dialing from %s to %d\n", t.tcpListener.Addr(), port)
	conn, err := net.DialTCP("tcp", nil,
		&net.TCPAddr{
			IP:   net.ParseIP("localhost"),
			Port: port,
		})
	// log.Println("Dialed", conn.Remo	)
	if err != nil {
		slog.Error("ERR dial failed", "port", port, err)
		return err
	}
	peer := &TCPPeer{
		conn:       conn,
		outbound:   true,
		listenAddr: serverLAddr,
	}
	s.addPeer <- peer
	// fmt.Printf("%s added 127.0.0.1:%d\n", serverLAddr, port)

	// fmt.Printf("%s is sending handshake request to 127.0.0.1:%d\n", serverLAddr, port)
	return s.SendHandshake(peer)
}

func (s *Server) OnPeer(peer *TCPPeer) error {
	s.peers[peer.conn.RemoteAddr()] = peer

	slog.Info("peer added", "addr", peer.conn.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.addPeer:
			if err := s.handlePeer(peer); err != nil {
				slog.Error("ERR handle peer", "err", err)
			}
			s.peers[peer.conn.RemoteAddr()] = peer
		case peer := <-s.delPeer:
			delete(s.peers, peer.conn.RemoteAddr())
			slog.Info("peer deleted", "addr", peer.conn.RemoteAddr())
		case msg := <-s.msgChan:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("ERR handle message", "err", err)
			}
		}
	}
}

func (s *Server) sendPeerList(p *TCPPeer) error {
	peerList := &MessagePeerList{
		Peers: []string{},
	}

	msg := NewMessage(s.ListenAddr.String(), peerList)

	for _, peer := range s.peers {
		peerList.Peers = append(peerList.Peers, peer.listenAddr)
	}
	if len(peerList.Peers) == 0 {
		return nil
	}

	// slog.Info("Peer list", "to", p.conn.RemoteAddr(), "list", strings.Join(peerList.Peers, ""))

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Println("gob", err)
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessagePeerList:
		return s.handlePeerList(v)
	}
	return nil
}

func (s *Server) handlePeerList(pl MessagePeerList) error {
	for i := 0; i < len(pl.Peers); i++ {
		// extract the port from the address
		fmt.Printf("peerlist => %s\n", pl.Peers)
		_, p, err := net.SplitHostPort(pl.Peers[i])
		if err != nil {
			slog.Error("ERR failed splitting", "err", err)
			return err
		}

		// convert port to int
		port, errr := strconv.Atoi(p)
		if errr != nil {
			slog.Error("ERR int to str conv", "err", errr)
		}

		if err := s.Dial(port, s.GameVariant, s.GameVersion, s.ListenAddr.String()); err != nil {
			slog.Error("Err failed to dial peer", "At port", port)
		}
	}

	return nil
}

func (s *Server) handlePeer(peer *TCPPeer) error {
	hp, err := s.PerformHandshake(peer)
	if err != nil {
		log.Println("ERR perform handshake", "err", err)
	}

	// Handle the peer connection
	go func(t *TCPPeer) {
		// log.Println("Handling Peer")
		if err := s.Transport.HandlePeer(peer, s.GameVariant, s.GameVersion); err != nil {
			slog.Error("ERR handle peer", "err", err)
		}
	}(peer)

	// If the peer is not outbound, send a handshake
	if !peer.outbound {
		if err := s.SendHandshake(peer); err != nil {
			slog.Error("ERR send handshake", "err", err)
		}
		if err := s.sendPeerList(peer); err != nil {
			slog.Error("ERR send peerlist", "err", err, "to", peer.conn.RemoteAddr())
		}

	}

	slog.Info("Added new connected", "peer", peer.conn.RemoteAddr(), "listenAddr", hp.ListenAddr,
		"variant", hp.GameVariant, "version", hp.Version, "we", s.ListenAddr)

	return nil
}

type HandshakePass struct {
	// GameStatus  GameStatus
	GameVariant GameVariant
	Version     string
	ListenAddr  string
}

func (s *Server) PerformHandshake(p *TCPPeer) (*HandshakePass, error) {
	hp := &HandshakePass{}
	if err := gob.NewDecoder(p.conn).Decode(hp); err != nil {
		return nil, err
	}

	// fmt.Printf("Peer listen addr: %s\n", p.listenAddr)

	if hp.GameVariant != s.GameVariant {
		return nil, fmt.Errorf("Game variant mismatch: %v != %v", hp.GameVariant, s.GameVariant)
	}

	if hp.Version != s.GameVersion {
		return nil, fmt.Errorf("Game version mismatch: %v != %v", hp.Version, s.GameVersion)
	}

	p.listenAddr = hp.ListenAddr

	// fmt.Printf("%s has successfully performed handshake with %s\n", addr, p.conn.RemoteAddr())
	return hp, nil
}

func (s *Server) SendHandshake(p *TCPPeer) error {
	hp := &HandshakePass{
		GameVariant: s.GameVariant,
		Version:     s.GameVersion,
		ListenAddr:  s.ListenAddr.String(),
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hp); err != nil {
		slog.Error("ERR send handshake", err, p.conn.RemoteAddr())
	}

	return p.Send(buf.Bytes())
}

func init() {
	gob.Register(MessagePeerList{})
	gob.Register(Message{})
}
