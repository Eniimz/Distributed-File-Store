package p2p

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

type Handshake func(Peer) error

func NOPHandshakeFunc(Peer) error { return nil }
