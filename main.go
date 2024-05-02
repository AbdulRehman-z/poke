package main

import (
	"net"
	p2p "poker/p2p"
)

func main() {
	s := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		Laddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 3000,
		},
	})
	if err := s.ListenAndAccept(); err != nil {
		panic(err)
	}
	select {}
}
