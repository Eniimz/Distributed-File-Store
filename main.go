package main

import (
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

	s.Start(t)

	return s
}

func main() {

	_ = makeServer(":3000")
	_ = makeServer(":4000", ":3000")

}
