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
		APIListenAddr: apiAddr,
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
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:3001/ready")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/ready")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/ready")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:6001/ready")

		// [3000:D, 4000:sb, 5000:bb, 7000]
		// PREFLOP
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:6001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:3001/fold")

		// FLOP
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:6001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:3001/fold")

		// TURN
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:6001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:3001/fold")

		// RIVER
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:6001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:3001/fold")

	}()

	time.Sleep(time.Millisecond * 200)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(time.Millisecond)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(time.Millisecond * 200)
	playerD.Connect(playerC.ListenAddr)

	select {}
}
