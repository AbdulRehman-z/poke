package p2p

import (
	"log"
	"log/slog"
	"net"
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
	// OnPeer:
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

func (s *Server) OnPeer(peer Peer) error {
	// s.peers[peer.RemoteAddr()] = peer

	// peer.Send([]byte("You are added!"))
	// slog.Info("peer added", "addr", peer.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.Transport.AddPeer():
			// slog.Info("peer added", "peer", peer.RemoteAddr(), "from", s.Transport.Addr())
			// if err := s.SendHandshake(peer); err != nil {
			// 	log.Println("ERR send handshake", "err", err)
			// }
			go SendHandshake(peer, s.GameVariant, s.GameVersion)
			// time.Sleep(2 * time.Second)
			if err := PerformHandshake(peer, s.GameVariant, s.GameVersion); err != nil {
				log.Println("ERR perform handshake", "err", err)
			}

			log.Println("Handshake sent")
			s.peers[peer.conn.RemoteAddr()] = peer
			// slog.Info("peer added", "peer", peer.RemoteAddr(), "to", s.Transport.Addr())
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
