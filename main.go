package main

import (
	"bufio"
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

	s1 := makeServer(":3001")                   // 0 bootstrap nodes (hub)
	s2 := makeServer(":4001", ":3001")          // 1 bootstrap node
	s3 := makeServer(":5001", ":3001", ":4001") // 2 bootstrap nodes
	s4 := makeServer(":6001", ":3001", ":5001") // 3 bootstrap nodes
	// s5 := makeServer(":7001", ":3001", ":5001", ":6001") // 4 bootstrap nodes

	go func() {
		if err := s1.Start(); err != nil {
			fmt.Printf("Server s1 error: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	go func() {
		if err := s2.Start(); err != nil {
			fmt.Printf("Server s2 error: %v\n", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	go func() {
		if err := s3.Start(); err != nil {
			fmt.Printf("Server s3 error: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
	go func() {
		if err := s4.Start(); err != nil {
			fmt.Printf("Server s4 error: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
	// go func() {
	// 	if err := s5.Start(); err != nil {
	// 		fmt.Printf("Server s5 error: %v\n", err)
	// 	}
	// }()
	// time.Sleep(500 * time.Millisecond)

	time.Sleep(5 * time.Second)

	// Print peer counts for all servers before storing
	fmt.Printf("\n=== PEER COUNTS BEFORE STORING ===\n")
	servers := []*FileServer{s1, s2, s3, s4}
	serverNames := []string{"s1", "s2", "s3", "s4"}

	for i, server := range servers {
		fmt.Printf("%s (%s): %d peers\n", serverNames[i], server.Transport.Addr(), len(server.peers))
	}
	fmt.Printf("========================================\n")

	data := bytes.NewReader([]byte("The Proposal data, yes"))
	key := "proposal.pdf"
	fmt.Printf("Storing file from s2...\n")
	if err := s2.StoreData(key, data); err != nil {
		fmt.Printf("The error while storing: %s", err)
	}

	// Wait a bit for connections to establish
	time.Sleep(5 * time.Second)

	fmt.Printf("\n=== All servers started ===\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  'peers' - Print all peer maps\n")
	fmt.Printf("  'exit'  - Shutdown and exit\n")
	fmt.Printf("> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		switch input {
		case "peers":
			fmt.Printf("\n=== CURRENT PEER MAPS ===\n")
			s1.PrintPeers()
			s2.PrintPeers()
			s3.PrintPeers()
			s4.PrintPeers()

		case "exit":
			goto shutdown
		default:
			fmt.Printf("Unknown command: %s\n", input)
		}

		fmt.Printf("> ")
	}

shutdown:

	fmt.Printf("\n=== FINAL PEER MAPS ===\n")
	s1.PrintPeers()
	s2.PrintPeers()
	s3.PrintPeers()
	s4.PrintPeers()

	fmt.Printf("Shutting down servers...\n")
	s1.Stop()
	s2.Stop()
	s3.Stop()
	s4.Stop()
}
