package main

import (
	"fmt"

	"github.com/eniimz/cas/p2p"
)

func main() {

	listenAdder := ":4000"

	opts := p2p.TCPTransportOpts{
		ListenAddress: listenAdder,
		HanshakeFunc:  p2p.NOPHandshakeFunc,
		Decoder:       p2p.NOPDecoder{},
	}

	tr := p2p.NewTCPTransport(opts)

	tr.OnPeer = func(p p2p.Peer) error {
		fmt.Println("The new peer connected")
		p.Close()
		return nil
	}

	tr.ListenAndAccept()

	for msg := range tr.Consume() {
		fmt.Printf("The message: %+v\n", msg)
	}

	select {}
}
