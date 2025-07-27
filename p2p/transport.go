package p2p

// The mode of communication that we are using
// which is transport
type Transport interface {
	ListenAndAccept() error

	// returns a channel of type RPC struct, which can be read
	//in a for loop
	Consume() <-chan Message
	Dial() error
}

type Peer interface {
}
