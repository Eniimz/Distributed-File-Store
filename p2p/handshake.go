package p2p

import (
	"fmt"
)

//Are we going to add things (properties) to this hanshake
//to this handshake struct, if we are to not add fields, better to
// use a singular func type and NOT this

// type Handshaker interface {
// 	Handshake() error
// }

// type NOPHandshake struct {}

// func (h NOPHandshake) Handshake() error {
// 	return nil
// }

//cant use this as func in a for loop
// var ErrorInvalidHandshake = errors.New("invalid handshake")

type Handshake func(Peer, string) error

func NOPHandshakeFunc(Peer, string) error { return nil }

func AddressExchangeHandshakeFunc(peer Peer, nodeId string) error {

	fmt.Printf("starting handshake by sending nodeId (%s) for node %s\n", nodeId, peer.RemoteAddr())
	if err := peer.Send([]byte(nodeId)); err != nil {
		fmt.Printf("The handshake error: %s\n", err)
		return err
	}
	fmt.Printf("Reading now...\n")
	buf := make([]byte, 1028)
	n, err := peer.Read(buf)
	if err != nil {
		fmt.Printf("Error while reading the remote nodeId: %s\n", err)
		return err
	}

	fmt.Printf("The no of bytes read for handshake from node %s: %s\n", peer.RemoteAddr(), string(buf[:n]))

	return nil
}
