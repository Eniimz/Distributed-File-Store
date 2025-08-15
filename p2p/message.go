package p2p

const (
	IncomingMessage = 0x01
	IncomingStream  = 0x02
)

type Message struct {
	Payload      []byte
	From         string
	RemotePeerId string
	Stream       bool
}
