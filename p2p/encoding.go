package p2p

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type NOPDecoder struct{}

func (d NOPDecoder) Decode(r io.Reader, msg *Message) error {

	reader := bufio.NewReader(r)

	buf := make([]byte, 2000)

	//this line blocks the thread and waits for data to be received and read
	//reader is a *bufio.Reader, wrapping r, which is net.Conn
	n, err := reader.Read(buf)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Read %d byes from the network", time.Now().Format("15:04:05.000"), n)

	msg.Payload = buf[:n]

	return nil
}

//here i was using earlier to read till delimeter line, err := reader.ReadBytes('\n')
//which has issues
//1. id data doesnt have \n at end, eof error
//2. even if it does, the go encoding coeeupts the data
//so for small files, we can wset a buffer of some space
//but for large file, we'll have to stream them
