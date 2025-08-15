package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcpTransport(t *testing.T) {

	listenAdder := ":4000"

	opts := TCPTransportOpts{
		ListenAddress: listenAdder,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       NOPDecoder{},
	}

	tr := NewTCPTransport(opts)

	assert.Equal(t, tr.ListenAddress, listenAdder)
	assert.Nil(t, tr.ListenAndAccept())

}
