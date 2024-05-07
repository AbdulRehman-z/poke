package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"strings"
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
	// Handler Handler

	// mu    sync.RWMutex
	peers map[net.Addr]*TCPPeer
}

func NewServer(opts *ServerConfig) *Server {
	return &Server{
		ServerConfig: opts,
		// Handler:      &DefaultHandler{},
		peers: make(map[net.Addr]*TCPPeer),
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

	// peer.Send([]byte("You are added!"))
	slog.Info("peer added", "addr", peer.conn.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.Transport.AddPeer():
			if err := PerformHandshake(peer, s.GameVariant, s.GameVersion, s.Transport.Addr()); err != nil {
				log.Println("ERR perform handshake", "err", err)
				// peer.conn.Close()
				continue
			}

			// Handle the peer connection
			go func(t *TCPPeer) {
				log.Println("Handling Peer")
				if err := s.Transport.HandlePeer(peer, s.GameVariant, s.GameVersion); err != nil {
					slog.Error("ERR handle peer", "err", err)
					// return err
				}
			}(peer)

			time.Sleep(1 * time.Second)

			// If the peer is not outbound, send a handshake
			if !peer.outbound {
				// log.Println("Sending Handshake")
				if err := SendHandshake(peer, s.GameVariant, s.GameVersion); err != nil {
					slog.Error("ERR send handshake", "err", err)
					continue
				}
				// time.Sleep( * time.Second)

				if err := s.sendPeerList(peer); err != nil {
					slog.Error("ERR send peerlist", "err", err, "to", peer.conn.RemoteAddr())
					continue
				}

			}

			s.peers[peer.conn.RemoteAddr()] = peer
		case peer := <-s.Transport.DelPeer():
			delete(s.peers, peer.conn.RemoteAddr())
			slog.Info("peer deleted", "addr", peer.conn.RemoteAddr())
		case msg := <-s.Transport.Consume():
			if err := s.handleMessage(msg); err != nil {
				slog.Error("ERR handle message", "err", err)
			}
		}
	}
}

func (s *Server) sendPeerList(p *TCPPeer) error {
	peerList := &MessagePeerList{
		Peers: make([]string, len(s.peers)),
	}

	msg := NewMessage(s.Transport.Addr(), peerList)

	i := 0
	for peer := range s.peers {
		peerList.Peers[i] = peer.String()
		i++
	}

	slog.Info("Peer list", "to", p.conn.RemoteAddr(), "list", strings.Join(peerList.Peers, ""))

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
		// case Message:
		// 	return nil
	}
	return nil
}

func (s *Server) handlePeerList(pl MessagePeerList) error {
	// i := 0
	// for _, peer := range pl.Peers {
	// 	//extract the port from the address
	// 	_, p, err := net.SplitHostPort(peer)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// convert port to int
	// 	port, error := strconv.Atoi(p)
	// 	if error != nil {
	// 		slog.Error("ERR int to str conv", "err", error)
	// 	}
	// 	if err := s.Transport.Dial(port, s.GameVariant, s.GameVersion); err != nil {
	// 		slog.Error("Err failed to dial peer", "At port", port)
	// 	}
	// 	i++
	// }

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

		log.Println(port)

		// panic("adwada")

		if err := s.Transport.Dial(port, s.GameVariant, s.GameVersion); err != nil {
			slog.Error("Err failed to dial peer", "At port", port)
		}
	}

	return nil
}

func init() {
	gob.Register(MessagePeerList{})
	gob.Register(Message{})
}
