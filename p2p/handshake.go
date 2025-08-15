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

type Message struct {
	Payload any
}

type Handshake func(Peer, MetaData) (string, error)

func NOPHandshakeFunc(Peer, MetaData) error { return nil }

func AddressExchangeHandshakeFunc(peer Peer, metadata MetaData) (string, error) {

	if err := sendHandshake(metadata, peer); err != nil {
		return "", err
	}

	data, err := receiveHandshake(metadata, peer)
	if err != nil {
		return "", err
	}

	return data, nil
}

func sendHandshake(metadata MetaData, peer Peer) error {

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

func receiveHandshake(metadata MetaData, peer Peer) (string, error) {

	var m MetaData

	if err := gob.NewDecoder(bytes.NewReader(metadata)).Decode(&m); err != nil {
		fmt.Printf("The error: %s\n", err)
	}

	return string(buf[:n]), nil
}
