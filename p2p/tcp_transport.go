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
	listenAddress string //opts
	listener      net.Listener
	shakeHands    Handshake //opts
	Decoder       Decoder   //opts
	rpcch         chan Message
}

func NewTCPTransport(listenAdder string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAdder,
		shakeHands:    NOPHandshakeFunc,
		Decoder:       NOPDecoder{},
		rpcch:         make(chan Message),
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

func (t *TCPTransport) Consume() <-chan Message {
	return t.rpcch
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

	//peer here is of type *TCPPeer struct
	//this Handshake func expects a param that implements Peer interface
	err := t.shakeHands(peer)
	if err != nil {
		conn.Close()
		fmt.Printf("TCP Handshake error: ", err)
		return
	}

	fmt.Printf("Handling the connection..")

	msg := Message{}
	for {

		if err := t.Decoder.Decode(conn, &msg); err != nil {
			fmt.Printf("TCP decode Error: %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr()

		//passing the  rpc to the channel
		// t.rpcch <- msg // have to pass the value of the struct not pointer

		fmt.Printf("The message: %v", msg)
	}

}
