package p2p

import (
	"bufio"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type NOPDecoder struct{}

func (d NOPDecoder) Decode(r io.Reader, msg *Message) error {

	reader := bufio.NewReader(r)

	// buf := make([]byte, 2000)

	// n, err := reader.Read(buf)

	line, err := reader.ReadBytes('\n')

	if err != nil {

		return err
	}

	// msg.Payload = buf[:n]
	msg.Payload = line

	return nil
}
