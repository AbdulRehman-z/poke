package main

import (
	"log"
	"net"
	"poker/deck"
	"poker/p2p"
	"poker/server"
	"time"
)

func initServer(version string, variant deck.GameVariant, port int) *server.Server {
	t := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		Laddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		},
	})

	opts := &server.ServerConfig{
		Transport:   t,
		GameVariant: variant,
		GameVersion: version,
	}

	return server.NewServer(opts)
}

func main() {
	s1 := initServer("GGPOKE V0.1-alpha", deck.TexasHoldings, 3000)
	s2 := initServer("GGPOKE V0.1-alpha", deck.TexasHoldings, 4000)
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
