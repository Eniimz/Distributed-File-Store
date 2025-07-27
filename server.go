package main

import (
	"fmt"
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
	Transport         p2p.TCPTransport
	bootstrappedNodes []string
	quitch            chan struct{}
	peerLock          sync.Mutex
	peers             map[string]p2p.Peer
	onPeer            func(p2p.Peer) error
}

func NewFileServer(opts *p2p.TCPTransport, nodes []string) *FileServer {

	return &FileServer{
		pathTransFormFunc: store.DefaultPathTransformFunc,
		Root:              store.DefaultRootName,
		Transport:         *opts,
		bootstrappedNodes: nodes,
		quitch:            make(chan struct{}),
		// onPeer:            func(p p2p.Peer) error { return nil },
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func consumeOrCloseLoop(s *FileServer) {

	defer func() {
		fmt.Printf("The server stop due to exit signal or some error")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Printf("The rpc received from the read loop: %+v", msg)
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
			fmt.Printf("\nConnectnig to remote peer: %s\n", addr)
			if err := s.Transport.Dial(addr); err != nil {
				fmt.Printf("Eror while connecting to remote peer: %s", err)
			}

		}(addr)

	}

}

func onPeer() {

}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fmt.Printf("\nListening on the port: %s", s.Transport.ListenAddress)

	//for each connection that is accepted, we check the bootstapped nodes len,
	//if > 0, then we dial all those nodes

	if len(s.bootstrappedNodes) > 0 {
		s.bootstrap(s.bootstrappedNodes)
	}

	consumeOrCloseLoop(s)

	return nil
}
