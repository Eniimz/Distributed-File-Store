package p2p

import (
	"fmt"
	"net"
)

type TCPPeer struct {
	conn net.Conn

	//outbound = true => dialing and retrieving the conn(outgoing)
	//outbound = false => accepting and retieving the connection (incoming)
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    Handshake
	Decoder       Decoder
}

func NewTCPTransport(listenAdder string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAdder,
		shakeHands:    NOPHandshakeFunc,
		Decoder:       NOPDecoder{},
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	// log.Println(t.listener)

	go startAcceptLoop(t)

	return nil
}

func startAcceptLoop(t *TCPTransport) {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}

}

func (t *TCPTransport) handleConn(conn net.Conn) {

	peer := NewTCPPeer(conn, true)

	err := t.shakeHands(peer)
	if err != nil {
		fmt.Printf("TCP Handshake error: ", err)
		return
	}

	fmt.Printf("Handling the connection..")

	// buf := make([]byte, 1028)

	// for {
	// 	n, err := conn.Read(buf)
	// 	if err != nil {
	// 		fmt.Printf("Buffer read error: %s", err)
	// 	}
	// 	fmt.Printf("The buffer sent: %s", buf[:n])
	// }
	for {
		t.Decoder.Decode(conn)
	}

}
