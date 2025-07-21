package p2p

import (
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader) error
}

type NOPDecoder struct{}

type Message struct {
	payload []byte
}

func (d NOPDecoder) Decode(r io.Reader) error {

	buf := make([]byte, 4026)

	msgNew := &Message{}

	n, err := r.Read(buf)
	if err != nil {
		return nil
	}

	msgNew.payload = buf[:n]

	fmt.Printf("The Message : %s", msgNew)

	return nil
}
