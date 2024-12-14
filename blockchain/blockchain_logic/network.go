package blockchain_logic

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type MessageType string

const (
	MessageTypeNewBlock           MessageType = "NEW_BLOCK"
	MessageTypeNewTx              MessageType = "NEW_TRANSACTION"
	MessageTypeBlockchain         MessageType = "BLOCKCHAIN_REQUEST"
	MessageTypeBlockchainResponse MessageType = "BLOCKCHAIN_RESPONSE"
	MessageTypeIPFSBackup         MessageType = "IPFS_BACKUP" // New message type
)

// BlockchainMessage represents a network message with blockchain-specific content
type BlockchainMessage struct {
	Type    MessageType `json:"type"`
	Content interface{} `json:"content"`
	From    string      `json:"from"`
	To      string      `json:"to,omitempty"`
}

// PeerConnection represents a connection to a peer
type PeerConnection struct {
	Address string
	Conn    net.Conn
}

// PeerNetwork manages peer connections and message broadcasting
type PeerNetwork struct {
	MyAddress   string
	Peers       map[string]*PeerConnection
	mutex       sync.RWMutex
	isConnected map[string]bool
	blockchain  *Blockchain // Reference to the blockchain
}

// NewPeerNetwork creates a new peer network
func NewPeerNetwork(myAddress string) *PeerNetwork {
	return &PeerNetwork{
		MyAddress:   myAddress,
		Peers:       make(map[string]*PeerConnection),
		isConnected: make(map[string]bool),
	}
}

// ConnectToPeersWithRetry establishes connections to other peers with retry mechanism
func (pn *PeerNetwork) ConnectToPeersWithRetry(peerAddresses []string, maxRetries int) {
	for _, addr := range peerAddresses {
		if addr == pn.MyAddress {
			continue // Skip self
		}

		go func(address string) {
			retryCount := 0
			for {
				pn.mutex.RLock()
				isConnected := pn.isConnected[address]
				pn.mutex.RUnlock()

				if isConnected {
					time.Sleep(5 * time.Second)
					continue
				}

				conn, err := net.Dial("tcp", address)
				if err != nil {
					retryCount++
					if retryCount <= maxRetries {
						fmt.Printf("Failed to connect to peer %s (attempt %d/%d): %v\n",
							address, retryCount, maxRetries, err)
						time.Sleep(5 * time.Second)
						continue
					}
					fmt.Printf("Gave up connecting to peer %s after %d attempts\n",
						address, maxRetries)
					break
				}

				pn.mutex.Lock()
				pn.Peers[address] = &PeerConnection{
					Address: address,
					Conn:    conn,
				}
				pn.isConnected[address] = true
				pn.mutex.Unlock()

				fmt.Printf("Successfully connected to peer: %s\n", address)

				// Start handling messages from this peer
				go pn.handleMessages(conn)
				break
			}
		}(addr)
	}
}

// StartServer starts listening for incoming connections
func (pn *PeerNetwork) StartServer() {
	listener, err := net.Listen("tcp", pn.MyAddress)
	if err != nil {
		fmt.Printf("Failed to start server on %s: %v\n", pn.MyAddress, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server started on %s\n", pn.MyAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go pn.handleConnection(conn)
	}
}

// handleConnection handles incoming peer connections
func (pn *PeerNetwork) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New connection from: %s\n", remoteAddr)

	pn.mutex.Lock()
	if _, exists := pn.Peers[remoteAddr]; !exists {
		pn.Peers[remoteAddr] = &PeerConnection{
			Address: remoteAddr,
			Conn:    conn,
		}
		pn.isConnected[remoteAddr] = true
	}
	pn.mutex.Unlock()

	go pn.handleMessages(conn)
}

// handleMessages handles incoming messages from a peer
func (pn *PeerNetwork) handleMessages(conn net.Conn) {
	defer func() {
		conn.Close()
		addr := conn.RemoteAddr().String()
		pn.mutex.Lock()
		delete(pn.Peers, addr)
		pn.isConnected[addr] = false
		pn.mutex.Unlock()
		fmt.Printf("Connection closed with peer: %s\n", addr)
	}()

	decoder := json.NewDecoder(conn)
	for {
		var message BlockchainMessage
		if err := decoder.Decode(&message); err != nil {
			fmt.Printf("Error decoding message from %s: %v\n", conn.RemoteAddr(), err)
			return
		}

		pn.handleMessage(message, conn)
	}
}

// handleMessage processes different types of blockchain messages
func (pn *PeerNetwork) handleMessage(message BlockchainMessage, conn net.Conn) {
	switch message.Type {
	case MessageTypeNewBlock:
		if block, ok := message.Content.(*Block); ok {
			fmt.Printf("Received new block from %s with hash %s\n", message.From, block.Hash)
			// Validate and add block to blockchain
			if pn.blockchain != nil {
				if err := pn.blockchain.AddBlock(block); err != nil {
					fmt.Printf("Error adding received block: %v\n", err)
				} else {
					// Forward the block to other peers (flooding)
					pn.BroadcastNewBlock(block)
				}
			}
		}

	case MessageTypeNewTx:
		if tx, ok := message.Content.(*Transaction); ok {
			fmt.Printf("Received new transaction from %s\n", message.From)
			// Add transaction to pool and forward to other peers
			pn.BroadcastTransaction(tx)
		}

	case MessageTypeBlockchain:
		// Handle blockchain request
		if pn.blockchain != nil {
			response := BlockchainMessage{
				Type:    MessageTypeBlockchainResponse,
				Content: pn.blockchain,
				From:    pn.MyAddress,
				To:      message.From,
			}
			json.NewEncoder(conn).Encode(response)
		}

	case MessageTypeBlockchainResponse:
		// Handle received blockchain
		if blockchain, ok := message.Content.(*Blockchain); ok {
			fmt.Printf("Received blockchain from %s\n", message.From)
			// Validate and potentially update local blockchain
			if pn.blockchain == nil || len(blockchain.Blocks) > len(pn.blockchain.Blocks) {
				if blockchain.IsValid() {
					pn.blockchain = blockchain
				}
			}
		}

	case MessageTypeIPFSBackup:
		// Handle IPFS backup hash
		if hash, ok := message.Content.(string); ok {
			fmt.Printf("Received blockchain backup hash from %s: %s\n", message.From, hash)

			if pn.blockchain != nil {
				// Restore from IPFS and validate
				tempBlocks, err := pn.blockchain.ipfsHandler.RetrieveBlockchain(hash)
				if err != nil {
					fmt.Printf("Error retrieving blockchain from IPFS: %v\n", err)
					return
				}

				// Only restore if the received blockchain is longer
				if len(tempBlocks) > len(pn.blockchain.Blocks) {
					err = pn.blockchain.RestoreFromIPFS(hash)
					if err != nil {
						fmt.Printf("Error restoring blockchain from IPFS: %v\n", err)
						return
					}
					fmt.Printf("Successfully restored blockchain from IPFS hash: %s\n", hash)
				}
			}
		}
	}
}

// BroadcastNewBlock broadcasts a new block to all peers
func (pn *PeerNetwork) BroadcastNewBlock(block *Block) {
	message := BlockchainMessage{
		Type:    MessageTypeNewBlock,
		Content: block,
		From:    pn.MyAddress,
	}
	pn.BroadcastMessage(string(message.Type), message)
}

// BroadcastTransaction broadcasts a new transaction to all peers
func (pn *PeerNetwork) BroadcastTransaction(tx *Transaction) {
	message := BlockchainMessage{
		Type:    MessageTypeNewTx,
		Content: tx,
		From:    pn.MyAddress,
	}
	pn.BroadcastMessage(string(message.Type), message)
}

// BroadcastMessage sends a message to all connected peers
func (pn *PeerNetwork) BroadcastMessage(messageType string, content interface{}) {
	pn.mutex.RLock()
	defer pn.mutex.RUnlock()

	for _, peer := range pn.Peers {
		go func(conn net.Conn) {
			if err := json.NewEncoder(conn).Encode(content); err != nil {
				fmt.Printf("Error broadcasting to %s: %v\n", conn.RemoteAddr(), err)
			}
		}(peer.Conn)
	}
}

// New method for broadcasting IPFS backup
func (pn *PeerNetwork) BroadcastIPFSBackup(hash string) {
	message := BlockchainMessage{
		Type:    MessageTypeIPFSBackup,
		Content: hash,
		From:    pn.MyAddress,
	}
	pn.BroadcastMessage(string(message.Type), message)
}

// SendToPeer sends a message to a specific peer
func (pn *PeerNetwork) SendToPeer(peerAddr string, messageType string, content interface{}) error {
	pn.mutex.RLock()
	peer, exists := pn.Peers[peerAddr]
	pn.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not connected", peerAddr)
	}

	return json.NewEncoder(peer.Conn).Encode(content)
}

// GetConnectedPeers returns a list of connected peer addresses
func (pn *PeerNetwork) GetConnectedPeers() []string {
	pn.mutex.RLock()
	defer pn.mutex.RUnlock()

	peers := make([]string, 0, len(pn.Peers))
	for addr := range pn.Peers {
		peers = append(peers, addr)
	}
	return peers
}

// IsConnected checks if a specific peer is connected
func (pn *PeerNetwork) IsConnected(peerAddr string) bool {
	pn.mutex.RLock()
	defer pn.mutex.RUnlock()
	return pn.isConnected[peerAddr]
}

// SetBlockchain sets the blockchain reference
func (pn *PeerNetwork) SetBlockchain(blockchain *Blockchain) {
	pn.blockchain = blockchain
}
