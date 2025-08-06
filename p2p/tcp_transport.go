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

func (p *TCPPeer) CloseStream() {
	p.Wg.Done()
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

// This implements the Transport interface
// This shows the address of the node
func (t *TCPTransport) Addr() string {
	return t.ListenAddress
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

	// a Read loop to read messages (rpcs) that are received
	for {
		msg := Message{} //rpc
		err = t.Decoder.Decode(conn, &msg)
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf("\nTCP network conn closed Error: %s\n", err)
			return
		}
		if err != nil {
			fmt.Printf("TCP read Error: %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr().String()

		if msg.Stream {
			peer.Wg.Add(1)
			fmt.Printf("Waiting till the stream is done:\n")
			peer.Wg.Wait()
			fmt.Printf("The stream is done and completedi\n\n")
			continue
		}

		t.rpcch <- msg
		//when data is passed here,the consumer also runs of the same nodek
		//thus it receives this msg
	}

}

//first send a msg that tells about the file
//then a msg with the file content
