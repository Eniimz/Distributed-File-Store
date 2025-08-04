package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
	//outbound = true => dialing and retrieving the conn(outgoing)
	//outbound = false => accepting and retieving the connection (incoming)
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

// This implements the Transport interface
func (p *TCPTransport) Close() error {

	return p.listener.Close()
}

type TCPTransportOpts struct {
	ListenAddress string    //opts
	HanshakeFunc  Handshake //opts
	Decoder       Decoder   //opts
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	OnPeer func(Peer) error
	rpcch  chan Message
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan Message),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		fmt.Printf("Listen error: %s", err)
		return err
	}

	// log.Println(t.listener)

	go startAcceptLoop(t)

	return nil

}

// This Dial function implements the transport interface
func (t *TCPTransport) Dial(listenAddress string) error {

	conn, err := net.Dial("tcp", listenAddress)
	if err != nil {
		return err
	}

	//after dialing we also have to listen to that connection (peer)
	//so we can send data back and forth

	go t.handleConn(conn, true)

	return nil

}

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Write(data)
	return err
}

func (t *TCPTransport) Addr() net.Addr {
	return t.listener.Addr()
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

		fmt.Printf("\nNew incoming connection\n")
		// we handle each new connection inside a different go routine
		go t.handleConn(conn, false)
	}

}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {

	var err error

	// Without defer, i'd have to manually call conn.Close()
	// at every exit point (returns):
	defer func() {
		fmt.Printf("dropping peer connection: %v", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	//peer here is of type *TCPPeer struct
	//this Handshake func expects a param that implements Peer interface
	if err := t.HanshakeFunc(peer); err != nil {
		// conn.Close()
		fmt.Printf("TCP Handshake error: %s", err)
		return
	}

	//if user populates OnPeer, we pass the peer connection to the func in params
	//and invoke it, otherwise we dont

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil { //when some node makes a connection to this node
			fmt.Printf("Peer error occureed... ")
			return
		}
	}

	msg := Message{} //rpc

	// a Read loop to read messages (rpcs) that are received
	for {
		fmt.Printf("\nIm waiting on the peer connection to read on conn %s\n", conn.RemoteAddr().String())
		err = t.Decoder.Decode(conn, &msg)
		fmt.Printf("\nAfter lower Decode i'm on conn %s\n", conn.RemoteAddr().String())
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf("\nTCP network conn closed Error: %s\n", err)
			return
		}
		if err != nil {
			fmt.Printf("TCP read Error: %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr().String()

		//when data is passed here,the consumer also runs of the same nodek
		//thus it receives this msg
		peer.Wg.Add(1)
		fmt.Printf("\nWaiting till the stream is done: ")
		t.rpcch <- msg
		peer.Wg.Wait()
		fmt.Printf("\nThe stream is done and completed")
		// fmt.Printf("The message: %+v", msg)
	}

}

//first send a msg that tells about the file
//then a msg with the file content
