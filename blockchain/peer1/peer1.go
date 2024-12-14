// peer1.go
package main

import (
	"blockchain/blockchain_logic"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const BACKUP_INTERVAL = 5 * time.Minute

func main() {
	// Configure peer addresses
	myAddress := "localhost:9001"
	peerAddresses := []string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	}

	fmt.Printf("Starting peer node on %s...\n", myAddress)

	// Initialize the peer network
	network := blockchain_logic.NewPeerNetwork(myAddress)

	// Initialize the blockchain with ML validator and training file
	blockchain, err := blockchain_logic.NewBlockchain(4, "../transactions.csv")
	if err != nil {
		fmt.Printf("Error initializing blockchain with ML validator: %v\n", err)
		os.Exit(1)
	}
	network.SetBlockchain(blockchain)

	// Start the server first
	go network.StartServer()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Read transactions from CSV
	transactionsPath := "../transactions.csv"
	transactions, err := blockchain_logic.ReadTransactionsFromCSV(transactionsPath)
	if err != nil {
		fmt.Printf("Error reading transactions: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully loaded %d transactions\n", len(transactions))
	blockchain_logic.PrintTransactions(transactions)

	// Connect to other peers with retry
	fmt.Println("Connecting to peers...")
	network.ConnectToPeersWithRetry(peerAddresses, 10)

	// Start mining process in a separate goroutine
	go func() {
		for {
			// Validate transactions before creating block
			validatedTransactions := blockchain.ValidateTransactionsML(transactions)

			if len(validatedTransactions) > 0 {
				// Create a new block with validated transactions
				latestBlock := blockchain.GetLatestBlock()
				newBlock := blockchain_logic.CreateBlock(
					latestBlock.Index+1,
					validatedTransactions,
					latestBlock.Hash,
					blockchain.Difficulty,
				)

				// Try to add the block to the blockchain
				if err := blockchain.AddBlock(newBlock); err != nil {
					fmt.Printf("Error adding block: %v\n", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Broadcast the new block to all peers
				fmt.Printf("Broadcasting new block with hash: %s\n", newBlock.Hash)
				network.BroadcastNewBlock(newBlock)
			} else {
				fmt.Println("No valid transactions to mine")
			}

			// Wait before mining next block
			time.Sleep(10 * time.Second)
		}
	}()

	// Start periodic IPFS backup
	go func() {
		for {
			time.Sleep(BACKUP_INTERVAL)
			hash, err := blockchain.BackupToIPFS()
			if err != nil {
				fmt.Printf("Error backing up blockchain to IPFS: %v\n", err)
				continue
			}

			// Broadcast the backup hash to peers
			message := blockchain_logic.BlockchainMessage{
				Type:    "IPFS_BACKUP",
				Content: hash,
				From:    myAddress,
			}
			network.BroadcastMessage(string(message.Type), message)
		}
	}()

	// Connection status checker
	go func() {
		for {
			time.Sleep(10 * time.Second)
			connectedPeers := network.GetConnectedPeers()
			fmt.Printf("\nConnected peers: %v\n", connectedPeers)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nPeer 1 is running...")
	fmt.Println("Press Ctrl+C to shutdown")

	<-sigChan
	fmt.Println("\nShutting down peer 1...")
}
