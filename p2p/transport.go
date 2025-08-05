package p2p

import "net"

// The mode of communication that we are using
// which is transport
type Transport interface {
	ListenAndAccept() error

	// returns a channel of type RPC struct, which can be read
	//in a for loop
	Consume() <-chan Message
	Dial(string) error
	Close() error
	Addr() string
}

type Peer interface {
	net.Conn
	Send([]byte) error
}
