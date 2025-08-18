package p2p

import (
	"bytes"
	"encoding/gob"
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

type Handshake func(Peer, HandshakeData) (*PeerInfo, error)

func NOPHandshakeFunc(Peer, HandshakeData) error { return nil }

func AddressExchangeHandshakeFunc(peer Peer, metadata HandshakeData) (*PeerInfo, error) {

	if err := sendHandshake(metadata, peer); err != nil {
		fmt.Printf("The err: %s\n", err)
		return nil, err
	}

	peerInfo, err := receiveHandshake(peer)
	if err != nil {
		fmt.Printf("The err: %s\n", err)
		return nil, err
	}

	return peerInfo, nil
}

func sendHandshake(metadata HandshakeData, peer Peer) error {

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(metadata); err != nil {
		fmt.Printf("The err: %s\n", err)
		return err
	}

	if err := peer.Send(buf.Bytes()); err != nil {
		fmt.Printf("The handshake error: %s\n", err)
		return err
	}

	return nil
}

func receiveHandshake(peer Peer) (*PeerInfo, error) {

	var p PeerInfo
	buf := make([]byte, 1028)

	n, err := peer.Read(buf)
	if err != nil {
		fmt.Printf("The error: %s\n", err)
		return nil, err
	}

	buf = buf[:n]

	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&p); err != nil {
		fmt.Printf("The error: %s\n", err)
	}

	return &p, nil
}
