package main

import (
	"fmt"

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
}

func NewFileServer(opts *p2p.TCPTransport, nodes []string) *FileServer {

	return &FileServer{
		pathTransFormFunc: store.DefaultPathTransformFunc,
		Root:              store.DefaultRootName,
		Transport:         *opts,
		bootstrappedNodes: nodes,
	}
}

func consumeOrCloseLoop(s *FileServer) {

	for {
		select {
		case rpc := <-s.Transport.Consume():
			fmt.Printf("The message received : %+v", rpc)

		}
	}

}

func bootstrap(t *p2p.TCPTransport, bootstrapNodes []string) {

	for {

		for _, addr := range bootstrapNodes {

			go func(addr string) {
				fmt.Printf("\nConnectnig to reomte peer: %s\n", addr)
				if err := t.Dial(addr); err != nil {
					fmt.Printf("Eror while connecting to remote peer: %s", err)
				}

			}(addr)

		}

	}

}

func (s *FileServer) Start(t *p2p.TCPTransport) {

	if err := t.ListenAndAccept(); err != nil {
		fmt.Printf("Error occured while listening")
	}

	fmt.Printf("\nListening on the port: %s", t.ListenAddress)

	//for each connection that is accepted, we check the bootstapped nodes len,
	//if > 0, then we dial all those nodes

	fmt.Printf("\nTHE LEN OF BOOTSTRAPPED NODES: %d", len(s.bootstrappedNodes))

	if len(s.bootstrappedNodes) > 0 {
		bootstrap(t, s.bootstrappedNodes)
	}

	consumeOrCloseLoop(s)

}
