package p2p

import (
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type NOPDecoder struct{}

func (d NOPDecoder) Decode(r io.Reader, msg *Message) error {

	buf := make([]byte, 2000)

	n, err := r.Read(buf)
	if err != nil {
		return nil
	}

	msg.Payload = buf[:n]

	return nil
}
