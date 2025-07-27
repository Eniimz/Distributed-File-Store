package main

import (
	"fmt"
	"log"
	"time"

	"github.com/eniimz/cas/p2p"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {

	opts := p2p.TCPTransportOpts{
		ListenAddress: listenAddress,
		HanshakeFunc:  p2p.NOPHandshakeFunc,
		Decoder:       p2p.NOPDecoder{},
	}

	t := p2p.NewTCPTransport(opts)

	s := NewFileServer(t, nodes)

	return s
}

func main() {

	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()
	// s1.Start()
	go func() {
		time.Sleep(6 * time.Second)
		log.Println("Stopping server on :3000")
		s1.Stop()

		//getting nil in the logs after stopping why?
	}()

	if err := s2.Start(); err != nil {
		fmt.Printf("\nits this nibba")
	}
}
