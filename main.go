package main

import (
	"bytes"
	"log"
	"strings"
	"time"

	"github.com/eniimz/cas/p2p"
	"github.com/eniimz/cas/store"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {

	root := strings.ReplaceAll(listenAddress, ":", "") + "_network"
	transportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddress,
		HanshakeFunc:  p2p.NOPHandshakeFunc,
		Decoder:       p2p.NOPDecoder{},
	}

	t := p2p.NewTCPTransport(transportOpts)

	storeOpts := store.StoreOpts{
		PathTransFormFunc: store.CASPathTransformFunc,
		Root:              root,
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

	for i := 0; i < 1; i++ {
		s2.StoreData("myPrivateDate", bytes.NewReader([]byte("The big data file")))
		time.Sleep(time.Second * 1)
	}

	if err := s2.store.Delete("myPrivateDate"); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)

	s2.Read("myPrivateDate")

	select {}
}
