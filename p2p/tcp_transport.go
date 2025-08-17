package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/eniimz/cas/encryption"
)

type PeerInfo struct {
	ID          string
	ListenAddrs string
}

type TCPPeer struct {
	net.Conn
	//outbound = true => dialing and retrieving the conn(outgoing)
	//outbound = false => accepting and retieving the connection (incoming)
	outbound bool
	Wg       *sync.WaitGroup
	PeerInfo
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {

	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

// This implements the Peer interface
func (p *TCPPeer) SetPeerInfo(pInfo *PeerInfo) {
	p.PeerInfo = PeerInfo{
		ID:          pInfo.ID,
		ListenAddrs: pInfo.ListenAddrs,
	}
}

// This implements the Peer interface
func (p *TCPPeer) GetPeerInfo() PeerInfo {
	return p.PeerInfo
}

// This implements the Peer interface
func (p *TCPPeer) CloseStream() {
	p.Wg.Done()
}

// This implements the Transport interface
func (p *TCPTransport) Close() error {
	return p.listener.Close()
}

type TCPTransportOpts struct {
	ListenAddress string    //opts
	HandshakeFunc Handshake //opts
	Decoder       Decoder   //opts
	NodeID        string
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	OnPeer   func(Peer) error
	rpcch    chan Message
}

type HandshakeData struct {
	ID          string
	ListenAddrs string
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {

	if len(opts.NodeID) == 0 {
		opts.NodeID = encryption.GenerateId()
	}

	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan Message),
	}
}

// This implements the Transport interface
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		fmt.Printf("Listen error: %s", err)
		return err
	}

	metadata := HandshakeData{
		ID:          t.NodeID,
		ListenAddrs: t.ListenAddress,
	}
	// log.Println(t.listener)

	go startAcceptLoop(t, metadata)

	return nil

}

// This Dial function implements the transport interface
func (t *TCPTransport) Dial(listenAddress string) error {

	conn, err := net.Dial("tcp", listenAddress)
	if err != nil {
		return err
	}

	metadata := HandshakeData{
		ID:          t.NodeID,
		ListenAddrs: t.ListenAddress,
	}

	//after dialing we also have to listen to that connection (peer)
	//so we can send data back and forth
	fmt.Printf("Now handling the dialed connection\n")
	go t.handleConn(conn, true, metadata)

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

func startAcceptLoop(t *TCPTransport, metadata HandshakeData) {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		// we handle each new connection inside a different go routine
		go t.handleConn(conn, false, metadata)
	}

}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool, metadata HandshakeData) {

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
	peerInfo, err := t.HandshakeFunc(peer, metadata)
	if err != nil {
		fmt.Printf("TCP Handshake error: %s", err)
		return
	}

	peer.SetPeerInfo(peerInfo)
	fmt.Printf("i\nThe TCP Peer after setting : %+v\n", peer)

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

		msg.From = peer.ListenAddrs
		msg.RemotePeerId = peer.ID

		if msg.Stream {
			peer.Wg.Add(1)
			fmt.Printf("Waiting till the stream is done\n")
			peer.Wg.Wait()
			fmt.Printf("The stream is done and completed\n")
			continue
		}

		t.rpcch <- msg
		//when data is passed here,the consumer also runs of the same nodek
		//thus it receives this msg
	}

}

//first send a msg that tells about the file
//then a msg with the file content
