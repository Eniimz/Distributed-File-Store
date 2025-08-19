package main

import (
	"bytes"
	"fmt"
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

	// Creating 5 servers

	s1 := makeServer(":3001")
	s2 := makeServer(":4001", ":3001")
	s3 := makeServer(":5001", ":3001", ":4001")
	s4 := makeServer(":6001", ":3001", ":5001")
	s5 := makeServer(":7001", ":3001", ":4001", ":5001")

	servers := []*FileServer{s1, s2, s3, s4, s5}

	// Start all servers
	for _, server := range servers {
		go func(s *FileServer) {
			if err := s.Start(); err != nil {
				fmt.Printf("Server error: %v\n", err)
			}
		}(server)
		time.Sleep(600 * time.Millisecond)
	}

	// Wait for servers to start
	time.Sleep(5 * time.Second)
	fmt.Printf("✅ All servers started successfully\n")

	// Print all servers' peers periodically
	go func() {
		time.Sleep(8 * time.Second) // Initial delay
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			fmt.Printf("\n📊 === ALL SERVERS PEER STATUS ===\n")
			for i, server := range servers {
				fmt.Printf("Server %d (%s): %d peers\n", i+1, server.Transport.Addr(), len(server.peers))
			}
			fmt.Printf("=====================================\n")
		}
	}()

	// Show peer status before starting file operations
	fmt.Printf("\n📊 === PEER STATUS BEFORE FILE STORAGE ===\n")
	for i, server := range servers {
		fmt.Printf("Server %d (%s): %d peers\n", i+1, server.Transport.Addr(), len(server.peers))
	}

	fmt.Printf("==========================================\n")

	reader, err := readFile("proposal.pdf")
	key := "proposal.pdf"
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	if err := s2.StoreData(key, reader); err != nil {
		fmt.Printf("Error storing file: %v\n", err)
		return
	}
	fmt.Printf("----File stored and distributed-----\n")

	// Wait for distribution
	time.Sleep(3 * time.Second)

	if err := s2.store.Delete(key, s2.NodeID); err != nil {
		fmt.Printf("Error deleting file: %s\n", err)
		return
	}

	_, err = s2.Read(key)
	if err != nil {
		fmt.Printf("Network retrieval failed: %s\n", err)
		return
	}
	fmt.Printf("File successfully retrieved from network\n")

	select {}
}
