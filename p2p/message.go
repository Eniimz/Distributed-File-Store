package p2p

import "net"

type Message struct {
	Payload []byte
	From    net.Addr
}
