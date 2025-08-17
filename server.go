package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/eniimz/cas/encryption"
	"github.com/eniimz/cas/p2p"
	"github.com/eniimz/cas/store"
)

// File Server listens for connections through our tcp
//peer discovery (bootstrapped networks)
//consume loop for data reading

type FileServer struct {
	pathTransFormFunc store.PathTransFormFunc
	Root              string
	bootstrappedNodes []string
	quitch            chan struct{}
	peerLock          sync.RWMutex
	peers             map[string]p2p.Peer
	EncKey            []byte
	NodeID            string
	Transport         p2p.Transport
	store             *store.Store
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key     string
	Size    int64
	OwnerID string
}

type MessageGetFile struct {
	Key     string
	OwnerID string
}

// ask for peers
type MessagePingPeers struct{}

// changing the map type here so that we can dial the peer
// receive the peers from the ping
type MessagePongPeers struct {
	Peers map[string]string
}

func NewFileServer(transportOpts *p2p.TCPTransport, nodes []string, storeOpts *store.Store, nodeId string) *FileServer {

	return &FileServer{
		pathTransFormFunc: store.DefaultPathTransformFunc,
		Root:              store.DefaultRootName,
		bootstrappedNodes: nodes,
		quitch:            make(chan struct{}),
		peers:             make(map[string]p2p.Peer),
		EncKey:            encryption.NewEncryptionKey(),
		NodeID:            nodeId,
		Transport:         transportOpts,
		store:             storeOpts,
	}
}

// as Peer interface implements net.Conn methods,
// for every peer strut we create a writer, and then multiWrite
// the payloas, that is send everyone the payload
func (s *FileServer) broadcast(p *Message) error {

	fmt.Printf("Broadcasting...\n")
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		fmt.Printf("The err: %s\n", err)
		return err
	}

	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	//writes to multple peers at once
	mw.Write([]byte{p2p.IncomingMessage})
	if _, err := mw.Write(buf.Bytes()); err != nil {
		fmt.Printf("The err: %s\n", err)
		return err
	}

	return nil
}

func (s *FileServer) Read(key string) (io.Reader, error) {

	if s.store.Has(key, s.NodeID) {
		fmt.Printf("The data exists in local disk, reading from the local disk...\n")
		_, r, err := s.store.Read(key, s.NodeID)
		if err != nil {
			fmt.Printf("The error: %s", err)
			return nil, err
		}

		if rc, ok := r.(io.ReadCloser); ok {
			fmt.Printf("Closing the reader\n")
			defer rc.Close()
		}

		return r, nil

	}

	fmt.Printf("[%s] data does not exist in local disk, reading from the network...\n", s.Transport.Addr())

	msg := Message{
		Payload: MessageGetFile{
			Key:     encryption.HashKey(key),
			OwnerID: s.NodeID,
		},
	}

	s.broadcast(&msg)

	// time.Sleep(time.Millisecond * 500)

	//after decrypting it
	for _, peer := range s.peers {

		fmt.Printf("\nreceiving peer from the network")

		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize), s.NodeID)
		if err != nil {
			fmt.Printf("Error in receiving the msg from remote peer %s: ", err)
			return nil, err
		}

		peer.CloseStream()
		fmt.Printf("received (%d) bytes from the network:\n", n)
	}

	_, r, err := s.store.Read(key, s.NodeID)
	if err != nil {
		fmt.Printf("The error: %s", err)
		return nil, err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Printf("Closing the reader\n")
		defer rc.Close()
	}

	return r, nil

}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	//store data into disk
	var (
		buf = new(bytes.Buffer)
		tee = io.TeeReader(r, buf)
	)

	n, err := s.store.Write(key, tee, s.NodeID)
	if err != nil {
		return err
	}

	msgBuf := Message{
		Payload: MessageStoreFile{
			Key:     encryption.HashKey(key),
			OwnerID: s.NodeID,
			// 16 bytes for the iv added by the encryptor
			Size: n + 16,
		},
	}

	s.broadcast(&msgBuf)

	//adding this timeSleep prevents the broadcasting, stream flag
	//and file data to be sent in a single stream

	//sending this stream flag after a delay allow
	time.Sleep(time.Millisecond * 6)

	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	//a warning message to the peer that the file is coming
	mw.Write([]byte{p2p.IncomingStream})
	//then send the file

	n, err = encryption.CopyEncrypt(s.EncKey, buf, mw)
	if err != nil {
		fmt.Printf("Error while encrypting the file: %s", err)
		return err
	}
	fmt.Printf("Encrypted %d bytes\n", n)

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

// PrintPeers prints all peers in the server's peer map
func (s *FileServer) PrintPeers() {
	fmt.Printf("DEBUG: PrintPeers called for server %s\n", s.Transport.Addr())
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	fmt.Printf("\n=== Peer Map for Server %s ===\n", s.Transport.Addr())
	if len(s.peers) == 0 {
		fmt.Printf("No peers connected\n")
		return
	}

	for peerID, peer := range s.peers {
		peerInfo := peer.GetPeerInfo()
		fmt.Printf("Peer ID: %s\n", peerID)
		fmt.Printf("  Listen Address: %s\n", peerInfo.ListenAddrs)
		fmt.Printf("  Remote Address: %s\n", peer.RemoteAddr())
		fmt.Printf("  Local Address: %s\n", peer.LocalAddr())
		fmt.Printf("---\n")
	}
	fmt.Printf("Total peers: %d\n\n", len(s.peers))
}

// The dialers read loop, but as well as the first read loop of the remote server that was dialed..
func (s *FileServer) consumeOrCloseLoop() {

	defer func() {
		fmt.Printf("The server stop due to exit signal or some error\n")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Printf("Received message from %+v\n", msg)
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				fmt.Printf("The error: %s\n", err)
			}
			if err := s.handleMessage(msg.From, msg.RemotePeerId, &m); err != nil {
				fmt.Printf("handle Message Error: %s", err)
			}
		case <-s.quitch:
			return
		}
	}

}

func (s *FileServer) handleMessage(from string, incomingPeerId string, msg *Message) error {

	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		if err := s.handleMessageStoreFile(from, incomingPeerId, v); err != nil {
			return err
		}
	case MessageGetFile:
		if err := s.handleMessageGetFile(from, incomingPeerId, v); err != nil {
			return err
		}
	case MessagePingPeers:
		if err := s.handleMessagePingPeers(from, incomingPeerId, v); err != nil {
			return err
		}
	case MessagePongPeers:
		if err := s.handleMessagePongPeers(from, incomingPeerId, v); err != nil {
			return err
		}

	}

	return nil
}

func (s *FileServer) handleMessagePingPeers(from string, incomingPeerID string, msg MessagePingPeers) error {

	fmt.Printf("Handlimg the ping peers message from: %s\n", from)
	fmt.Printf("The peers map:\n")
	for key, peer := range s.peers {
		fmt.Printf("Peer (%s):%+v \n", key, peer)
	}

	fmt.Printf("Removing the remote peer before sending new peera...\n")

	peers := make(map[string]string)
	for key, peer := range s.peers {
		if key != incomingPeerID {
			peers[key] = peer.GetPeerInfo().ListenAddrs
		}
	}

	if len(peers) == 0 {
		fmt.Printf("No peers to ping\n")
		return nil
	}

	fmt.Printf("The newly filtered peers...\n")
	for key, addr := range peers {
		fmt.Printf("Peer(%s): %s\n", key, addr)
	}

	pongMessage := &Message{
		Payload: MessagePongPeers{
			Peers: peers,
		},
	}

	fmt.Printf("Pinging the peers..\n")

	s.sendToPeer(incomingPeerID, pongMessage)

	return nil
}

func (s *FileServer) sendToPeer(senderId string, msg *Message) error {

	peer, ok := s.peers[senderId]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", peer)
	}

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		fmt.Printf("Error during encoding one peer msg: %s\n", err)
		return err
	}

	peer.Send([]byte{p2p.IncomingMessage})
	if err := peer.Send(buf.Bytes()); err != nil {
		fmt.Printf("Error in sending msg to specific peer: %s", err)
		return err
	}

	return nil

}

func (s *FileServer) handleMessagePongPeers(from string, incomingPeerID string, msg MessagePongPeers) error {

	_, ok := s.peers[incomingPeerID]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	receivedPeers := msg.Peers
	fmt.Printf("Dialing all the receved peers from the Pong message..")
	for _, addr := range receivedPeers {
		go func(addr string) {

			if err := s.Transport.Dial(addr); err != nil {
				fmt.Printf("Error while dialing received pong peer: %s", err)
			}

		}(addr)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, incomingPeerID string, msg MessageStoreFile) error {

	peer, ok := s.peers[incomingPeerID]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	_, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size), msg.OwnerID)
	if err != nil {
		return err
	}
	peer.CloseStream()

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, incomingPeerID string, msg MessageGetFile) error {

	if !s.store.Has(msg.Key, msg.OwnerID) {
		return fmt.Errorf("data couldnt be found in the remote peer: %s", from)
	}

	fileSize, r, err := s.store.Read(msg.Key, msg.OwnerID)
	if err != nil {
		return err
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		fmt.Printf("Closing the reader\n")
		defer rc.Close()
	}

	peer, ok := s.peers[incomingPeerID]
	if !ok {
		return fmt.Errorf("peer not found in the peers map: %s", from)
	}

	peer.Send([]byte{p2p.IncomingStream})

	binary.Write(peer, binary.LittleEndian, fileSize)
	_, err = io.Copy(peer, r)
	if err != nil {
		fmt.Printf("error while writing the rest of the file to the network: %s", err)
		return err
	}

	return nil
}

func (s *FileServer) bootstrap(bootstrapNodes []string) {

	fmt.Printf("\nbootstraping the nodes...\n")
	for _, addr := range bootstrapNodes {

		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("Dialing remote peer: %s\n", addr)
			if err := s.Transport.Dial(addr); err != nil {
				fmt.Printf("Error while connecting to remote peer: %s", err)
			}

		}(addr)

	}

}

// to ensure server itself isnt included in the peers map
// its already done so by the design of this logic..
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()

	//in case of multiple conn accepted, peers map will be written to concurrently
	//peers map => not thread safe
	defer s.peerLock.Unlock()

	fmt.Printf("New peer connected: %+v\n", p.GetPeerInfo())
	s.peers[p.GetPeerInfo().ID] = p
	// peer{ remoteAddr : remoteAddr}

	return nil
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fmt.Printf("\nListening on the port: %s", s.Transport.Addr())

	//for each connection that is accepted, we check the bootstapped nodes len,
	//if > 0, then we dial all those nodes

	if len(s.bootstrappedNodes) > 0 {
		s.bootstrap(s.bootstrappedNodes)
	}

	go func() {

		time.Sleep(time.Second * 2)
		//pinging the remote peers of the just accepted conn
		fmt.Printf("The First ping to the remote peers.\n")
		msg := Message{
			Payload: MessagePingPeers{},
		}

		s.broadcast(&msg)
	}()

	s.consumeOrCloseLoop()

	return nil
}

// any type that is embedded into the type that is being encoded by gob
// needs to be registered here
func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessagePingPeers{})
	gob.Register(MessagePongPeers{})
}
