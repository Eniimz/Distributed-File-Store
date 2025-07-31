package main

import (
	"log"
	"strings"
	"time"

	"github.com/eniimz/cas/p2p"
	"github.com/eniimz/cas/store"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {

	transportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddress,
		HanshakeFunc:  p2p.NOPHandshakeFunc,
		Decoder:       p2p.NOPDecoder{},
	}

	t := p2p.NewTCPTransport(transportOpts)

	storeOpts := store.StoreOpts{
		PathTransFormFunc: store.CASPathTransformFunc,
		Root:              store.DefaultRootName,
	}

	store := store.NewStore(storeOpts)

	s := NewFileServer(t, nodes, store)

	t.OnPeer = s.OnPeer

	return s
}

func main() {

	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(2 * time.Second)

	data := strings.NewReader("This is my big data file")

	s2.StoreData("myPrivateData", data)

	select {}
}

//11
