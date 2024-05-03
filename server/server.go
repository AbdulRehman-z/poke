package server

import (
	"log/slog"
	"net"
	"poker/p2p"
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
	GameVariant GameVariant
	GameVersion string
	Transport   p2p.Transport
}

type Server struct {
	*ServerConfig
	Handler p2p.Handler

	// mu    sync.RWMutex
	peers map[net.Addr]p2p.Peer
}

func NewServer(opts *ServerConfig) *Server {
	return &Server{
		ServerConfig: opts,
		Handler:      &p2p.DefaultHandler{},
		// msgCh:        make(chan *p2p.Message),
		// addPeerCh:    make(chan *p2p.TCPPeer),
		// delPeerCh:    make(chan *p2p.TCPPeer),
		// quitCh:       make(chan struct{}),
		peers: make(map[net.Addr]p2p.Peer),
	}
}

func (s *Server) Start() error {
	err := s.Transport.ListenAndAccept()
	if err != nil {
		slog.Error("ERR failed to listen", "err", err)
	}
	s.loop()
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.Transport.AddPeer():
			s.peers[peer.RemoteAddr()] = peer
			peer.Send([]byte("You are added!"))
			slog.Info("peer added", "addr", peer.RemoteAddr())
		case msg := <-s.Transport.Consume():
			if err := s.Handler.HandleMessage(msg); err != nil {
				slog.Error("ERR handle message", "err", err)
			}
		}
	}
}
