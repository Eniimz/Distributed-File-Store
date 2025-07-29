package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

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

type DataMessage struct {
	Key string
	//gob: type not registered for interface: bytes.Buffer
	//(Data io.Reader)  getting some gob error with an interface
	Data []byte
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

	tee := io.TeeReader(r, buf)

	if err := s.store.Write(key, tee); err != nil {
		return err
	}

	//why pass a pointer?
	//better to not create copies of stuff like data, files etc
	p := &DataMessage{
		Key:  key,
		Data: buf.Bytes(),
	}

	//broadcast the data to known peers in the network
	if err := s.broadcast(&Message{
		Payload: p.Data,
		From:    "todo",
	}); err != nil {
		log.Fatal(err)
		return err
	}

	// }

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) consumeOrCloseLoop() {

	defer func() {
		fmt.Printf("The server stop due to exit signal or some error")
		s.Transport.Close()
	}()

	for {

		select {
		case msg := <-s.Transport.Consume():
			fmt.Printf("\nThe recv msg: %+v", msg)
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				fmt.Printf("The error: %s", err)
			}
			fmt.Printf("\nThe msg received: %+v", string(m.Payload))
		case <-s.quitch:
			return
		}
	}

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

	log.Printf("Connected with remote peer %s", p.RemoteAddr())

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
