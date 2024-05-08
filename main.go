package main

import (
	"log"
	"net"
	"poker/p2p"
	"time"
)

func initServer(port int) *p2p.Server {
	// tcpTransportOpts := &p2p.TCPTransportOpts{
	// 	Laddr: &net.TCPAddr{
	// 		IP:   net.ParseIP("127.0.0.1"),
	// 		Port: port,
	// 	},
	// 	// Handshake: p2p.PerformHandshake,
	// }

	// t := p2p.NewTCPTransport(tcpTransportOpts)
	laddr := &net.TCPAddr{
		IP:   net.ParseIP("localhost"),
		Port: port,
	}

	opts := &p2p.ServerConfig{
		ListenAddr:  laddr,
		GameVariant: p2p.TexasHoldings,
		GameVersion: "GGPOKE V0.1-alpha",
	}

	s := p2p.NewServer(opts)
	return s
}

func main() {
	player1 := initServer(3000)
	player2 := initServer(4000)
	// player3 := initServer(5000)
	// player4 := initServer(6000)

	// start player1
	go func() {
		log.Fatal(player1.Start())
	}()

	time.Sleep(1 * time.Second)
	//start player2
	go func() {
		log.Fatal(player2.Start())
	}()

	// time.Sleep(1 * time.Second)
	// // start player3
	// go func() {
	// 	log.Fatal(player3.Start())
	// }()

	// time.Sleep(1 * time.Second)
	// start player4
	// go func() {
	// 	log.Fatal(player4.Start())
	// }()

	time.Sleep(1 * time.Second)
	// player2 connects with player1, and add player 1 to its peer list
	if err := player2.Dial(3000, p2p.GameVariant(player2.GameVariant), player2.GameVersion, player2.ListenAddr.String()); err != nil {
		log.Println(err)
	}

	time.Sleep(1 * time.Second)

	// player3 connects with player2, and add player1 and player2 to its peer list
	// if err := player3.Transport.Dial(4000, p2p.GameVariant(player2.GameVariant), player2.GameVersion, player2.ListenAddr); err != nil {
	// 	log.Println(err)
	// }

	// // player4 connects with player2, and add player1, player2 to its peer list
	// if err := player4.Transport.Dial(5000, p2p.GameVariant(player3.GameVariant), player3.GameVersion); err != nil {
	// 	log.Println(err)
	// }

	select {}
}
