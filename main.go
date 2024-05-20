package main

import (
	"net/http"
	"time"

	"github.com/AbdulRehman-z/poke/p2p"
)

func makeServerAndStart(addr string, apiAddr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:       "GGPOKER V0.1-alpha",
		ListenAddr:    addr,
		ApiListenAddr: apiAddr,
		GameVariant:   p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	playerA := makeServerAndStart(":3000", ":3001")
	playerB := makeServerAndStart(":4000", ":4001")
	playerC := makeServerAndStart(":5000", ":5001")
	playerD := makeServerAndStart(":6000", ":6001")

	go func() {
		time.Sleep(time.Second * 3)
		http.Get("http://localhost:3001/ready")

		time.Sleep(time.Second * 3)
		http.Get("http://localhost:4001/ready")
	}()

	time.Sleep(time.Millisecond * 200)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(time.Millisecond)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(time.Millisecond * 200)
	playerD.Connect(playerC.ListenAddr)

	select {}
}
