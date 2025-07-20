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
}

func NewTCPTransport(listenAdder string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAdder,
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
	fmt.Println("Handling the connection..: ", conn)
}
