package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/eniimz/cas/p2p"
	"github.com/eniimz/cas/store"
)

// File Server listens for connections through our tcp
//peer discovery (bootstrapped networks)
//consume loop for data reading

type FileServer struct {
	pathTransFormFunc store.PathTransFormFunc
	Root              string
	Transport         p2p.Transport
	bootstrappedNodes []string
	quitch            chan struct{}
	store             *store.Store
	peerLock          sync.Mutex
	peers             map[string]p2p.Peer
}

type Message struct {
	Payload []byte
	From    string
}

func NewFileServer(transportOpts *p2p.TCPTransport, nodes []string, storeOpts *store.Store) *FileServer {

	return &FileServer{
		pathTransFormFunc: store.DefaultPathTransformFunc,
		Root:              store.DefaultRootName,
		Transport:         transportOpts,
		bootstrappedNodes: nodes,
		quitch:            make(chan struct{}),
		store:             storeOpts,
		peers:             make(map[string]p2p.Peer),
		// onPeer:     	       func(p p2p.Peer) error { return nil },
	}
}

// as Peer interface implements net.Conn methods,
// for every peer strut we create a writer, and then multiWrite
// the payloas, that is send everyone the payload
func (s *FileServer) broadcast(p *Message) error {

	for _, peer := range s.peers {
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(p); err != nil {
			return err
		}

		if _, err := peer.Write(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileServer) StoreData(key string, r io.Reader) error {

	//store data into disk
	buf := new(bytes.Buffer)

	msg := Message{
		Payload: []byte(key),
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if _, err := peer.Write(buf.Bytes()); err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(2 * time.Second)

	payload := []byte("The large file content")

	for _, peer := range s.peers {
		if _, err := peer.Write(payload); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

// The dialers read loop, but as well as the first read loop of the remote server that was dialed..
func (s *FileServer) consumeOrCloseLoop() {

	defer func() {
		fmt.Printf("The server stop due to exit signal or some error")
		s.Transport.Close()
	}()

	for {
		fmt.Printf("\nIm waiting on the channel to read")
		select {
		case msg := <-s.Transport.Consume():
			fmt.Printf("\nThe recv msg: %+v", msg)

			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				fmt.Printf("The error: %s", err)
			}

			fmt.Printf("\nThe msg received: %s", string(m.Payload))

		case <-s.quitch:
			return
		}
	}

}

func (s *FileServer) handleMessage() {

}

func (s *FileServer) handleMessageStoreFile(from string) {

	buf := make([]byte, 1000)
	peer := s.peers[from]

	fmt.Printf("\nThe peer: %+v\n", peer)
	_, err := peer.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nThe file content: %s", string(buf))
}

func (s *FileServer) bootstrap(bootstrapNodes []string) {

	fmt.Printf("\nbootstraping the nodes...\n")
	for _, addr := range bootstrapNodes {

		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("\nConnecting to remote peer: %s\n", addr)
			if err := s.Transport.Dial(addr); err != nil {
				fmt.Printf("Error while connecting to remote peer: %s", err)
			}

		}(addr)

	}

}

// to ensure server itself isnt included in the peers map
// its already done so by the design of this logic..
func (s *FileServer) OnPeer(p p2p.Peer) error {
	// s.peerLock.Lock()

	// defer s.peerLock.Unlock()
	// peer{ remoteAddr : remoteAddr}
	s.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected with remote peer %+v", s.peers)

	return nil
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fmt.Printf("\nListening on the port: %s", s.Transport.Addr().String())

	//for each connection that is accepted, we check the bootstapped nodes len,
	//if > 0, then we dial all those nodes

	if len(s.bootstrappedNodes) > 0 {
		s.bootstrap(s.bootstrappedNodes)
	}

	s.consumeOrCloseLoop()

	return nil
}
