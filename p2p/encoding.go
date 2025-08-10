package p2p

import (
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type NOPDecoder struct{}

func (d NOPDecoder) Decode(r io.Reader, msg *Message) error {

	peakBuf := make([]byte, 1)
	_, err := r.Read(peakBuf)
	if err != nil {
		return err
	}

	msg.Stream = peakBuf[0] == IncomingStream
	if msg.Stream {
		return nil
	}

	buf := make([]byte, 1028)
	//this line blocks the thread and waits for data to be received and read
	//reader is a *bufio.Reader, wrapping r, which is net.Conn
	n, err := r.Read(buf)
	if err != nil {
		fmt.Printf("Error in reading the message: %s", err)
		return err
	}

	fmt.Printf("Received (%d) bytes from the network\n", n)

	msg.Payload = buf[:n]

	return nil
}

//here i was using earlier to read till delimeter line, err := reader.ReadBytes('\n')
//which has issues
//1. id data doesnt have \n at end, eof error
//2. even if it does, the go encoding coeeupts the data
//so for small files, we can wset a buffer of some space
//but for large file, we'll have to stream them
