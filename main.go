package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eniimz/cas/encryption"
	"github.com/eniimz/cas/p2p"
	"github.com/eniimz/cas/store"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {

	root := strings.ReplaceAll(listenAddress, ":", "") + "_network"
	nodeID := encryption.GenerateId()

	transportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.AddressExchangeHandshakeFunc,
		Decoder:       p2p.NOPDecoder{},
		NodeID:        nodeID,
	}

	t := p2p.NewTCPTransport(transportOpts)

	storeOpts := store.StoreOpts{
		PathTransFormFunc: store.CASPathTransformFunc,
		Root:              root,
	}

	store := store.NewStore(storeOpts)

	s := NewFileServer(t, nodes, store, nodeID)

	t.OnPeer = s.OnPeer

	return s
}

func readFile(filePath string) (io.Reader, error) {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf, nil
}

func main() {

	s1 := makeServer(":3001")
	s2 := makeServer(":4001", ":3001")
	s3 := makeServer(":5001", ":3001")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()
	// time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s3.Start())
	}()

	// time.Sleep(2 * time.Second)

	// for i := 0; i < 20; i++ {

	// key := fmt.Sprintf("myPrivateDate_%d", i)
	// data, err := readFile("proposal.pdf")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// key := "proposal.pdf"
	// s2.StoreData(key, data)

	// if err := s2.store.Delete(key); err != nil {
	// 	log.Fatal(err)
	// }

	// if _, err := s2.Read(key); err != nil {
	// 	log.Fatal(err)
	// }

	// }

	// r, err := readFile("proposal.pdf")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("This the read data of the file: ", r)

	select {}
}
