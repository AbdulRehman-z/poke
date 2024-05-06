package p2p

import (
	"log/slog"
	"net"
	"time"
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
	Transport   Transport
}

type Server struct {
	*ServerConfig
	Handler Handler

	// mu    sync.RWMutex
	peers map[net.Addr]*TCPPeer
}

func NewServer(opts *ServerConfig) *Server {
	return &Server{
		ServerConfig: opts,
		Handler:      &DefaultHandler{},
		peers:        make(map[net.Addr]*TCPPeer),
	}
}

func (s *Server) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.loop()
	return nil
}

func (s *Server) OnPeer(peer *TCPPeer) error {
	s.peers[peer.conn.RemoteAddr()] = peer

	peer.Send([]byte("You are added!"))
	slog.Info("peer added", "addr", peer.conn.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.Transport.AddPeer():
			go SendHandshake(peer, s.GameVariant, s.GameVersion)
			time.Sleep(1 * time.Second)
			go s.Transport.HandlePeer(peer, s.GameVariant, s.GameVersion)
			// s.peers[peer.conn.RemoteAddr()] = peer
			// slog.Info("peer added", "peer", peer.conn.RemoteAddr(), "to", s.Transport.Addr())
		case peer := <-s.Transport.DelPeer():
			delete(s.peers, peer.conn.RemoteAddr())
			slog.Info("peer deleted", "addr", peer.conn.RemoteAddr())
		case msg := <-s.Transport.Consume():
			if err := s.Handler.HandleMessage(msg); err != nil {
				slog.Error("ERR handle message", "err", err)
			}
		}
	}
}
