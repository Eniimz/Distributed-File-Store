package main

import (
	"github.com/eniimz/cas/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":4000")

	tr.ListenAndAccept()

	select {}
}
