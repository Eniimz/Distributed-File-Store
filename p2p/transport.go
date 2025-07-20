package p2p

// The mode of communication that we are using
// which is transport
type Transport interface {
	ListenAndAccept() error
}
