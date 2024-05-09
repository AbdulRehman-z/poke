package main

import (
	"time"

	"github.com/anthdm/ggpoker/p2p"
)

func makeServerAndStart(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:     "GGPOKER V0.1-alpha",
		ListenAddr:  addr,
		GameVariant: p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	playerA := makeServerAndStart(":3000")
	playerB := makeServerAndStart(":4000")
	playerC := makeServerAndStart(":5000")
	// playerD := makeServerAndStart(":6000")
	// playerE := makeServerAndStart(":7000")
	// playerF := makeServerAndStart(":8000")

	time.Sleep(time.Millisecond * 200)
	playerB.Connect(playerA.ListenAddr) // 1
	time.Sleep(time.Millisecond * 200)
	playerC.Connect(playerB.ListenAddr) // 2
	// time.Sleep(time.Millisecond * 200)
	// playerD.Connect(playerC.ListenAddr) // 3
	// time.Sleep(time.Millisecond * 200)
	// playerE.Connect(playerD.ListenAddr) // 4
	// time.Sleep(time.Millisecond * 200)
	// playerF.Connect(playerE.ListenAddr) // 5

	select {}
}
