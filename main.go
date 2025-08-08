package main

import (
	"bytes"
	"fmt"
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
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s3.Start())
	}()

	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {

		key := fmt.Sprintf("myPrivateDate_%d", i)
		s2.StoreData(key, bytes.NewReader([]byte("The big data file")))

		if err := s2.store.Delete(key); err != nil {
			log.Fatal(err)
		}

		if _, err := s2.Read(key); err != nil {
			log.Fatal(err)
		}

	}

	select {}
}
