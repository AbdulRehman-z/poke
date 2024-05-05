package main

import (
	"log"
	"net"
	"poker/p2p"
	"time"
)

func initServer(version string, variant p2p.GameVariant, port int) *p2p.Server {
	tcpTransportOpts := p2p.TCPTransportOpts{
		Laddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		},
		// Handshake: p2p.PerformHandshake,
	}

	t := p2p.NewTCPTransport(tcpTransportOpts)

	opts := &p2p.ServerConfig{
		Transport:   t,
		GameVariant: variant,
		GameVersion: version,
	}

	s := p2p.NewServer(opts)
	// t.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := initServer("GGPOKE V0.1-alpha", p2p.TexasHoldings, 3000)
	s2 := initServer("GGPOKE V0.1-alpha", p2p.TexasHoldings, 4000)
	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(1 * time.Second)
	if err := s2.Transport.Dial(3000); err != nil {
		log.Println(err)
	}
	select {}
}
