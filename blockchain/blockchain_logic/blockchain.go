// blockchain.go
package blockchain_logic

import (
	"fmt"
	"sync"
)

// Blockchain struct
type Blockchain struct {
	Blocks      []*Block
	mutex       sync.RWMutex
	Difficulty  int
	MLValidator *MLTransactionValidator
	ipfsHandler *IPFSHandler // Added IPFS handler
}

// Single NewBlockchain function that handles ML validator initialization
func NewBlockchain(difficulty int, trainingFile string) (*Blockchain, error) {
	validator := NewMLTransactionValidator()
	err := validator.Train(trainingFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ML validator: %v", err)
	}

	// Initialize IPFS handler
	ipfsHandler, err := NewIPFSHandler("localhost:5001")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize IPFS handler: %v", err)
	}

	blockchain := &Blockchain{
		Blocks:      make([]*Block, 0),
		Difficulty:  difficulty,
		MLValidator: validator,
		ipfsHandler: ipfsHandler,
	}

	// Create genesis block
	genesisBlock := CreateBlock(0, []Transaction{}, "", difficulty)
	blockchain.AddBlock(genesisBlock)

	return blockchain, nil
}

// Method to validate transactions using ML
func (bc *Blockchain) ValidateTransactionsML(transactions []Transaction) []Transaction {
	validTransactions := make([]Transaction, 0)

	for _, tx := range transactions {
		isValid, confidence, reason := bc.MLValidator.ValidateTransaction(tx)
		if isValid {
			validTransactions = append(validTransactions, tx)
			fmt.Printf("Transaction validated (confidence: %.2f%%): %s\n", confidence*100, reason)
		} else {
			fmt.Printf("Transaction rejected (confidence: %.2f%%): %s\n", confidence*100, reason)
		}
	}

	return validTransactions
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.Blocks) > 0 {
		currentBlock := block
		previousBlock := bc.Blocks[len(bc.Blocks)-1]

		if currentBlock.PrevHash != previousBlock.Hash {
			return fmt.Errorf("invalid previous hash")
		}

		if !currentBlock.ValidateBlock() {
			return fmt.Errorf("invalid block proof of work")
		}
	}

	// Store block in IPFS
	ipfsHash, err := bc.ipfsHandler.StoreBlock(block)
	if err != nil {
		return fmt.Errorf("failed to store block in IPFS: %v", err)
	}

	// Pin the block to ensure it's kept in the network
	if err := bc.ipfsHandler.Pin(ipfsHash); err != nil {
		return fmt.Errorf("failed to pin block in IPFS: %v", err)
	}

	fmt.Printf("Block stored in IPFS with hash: %s\n", ipfsHash)

	bc.Blocks = append(bc.Blocks, block)
	return nil
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		if currentBlock.PrevHash != previousBlock.Hash {
			return false
		}

		if !currentBlock.ValidateBlock() {
			return false
		}
	}
	return true
}

// New method to backup blockchain to IPFS
func (bc *Blockchain) BackupToIPFS() (string, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	hash, err := bc.ipfsHandler.StoreBlockchain(bc)
	if err != nil {
		return "", fmt.Errorf("failed to backup blockchain to IPFS: %v", err)
	}

	if err := bc.ipfsHandler.Pin(hash); err != nil {
		return "", fmt.Errorf("failed to pin blockchain backup: %v", err)
	}

	fmt.Printf("Blockchain backed up to IPFS with hash: %s\n", hash)
	return hash, nil
}

// New method to restore blockchain from IPFS
func (bc *Blockchain) RestoreFromIPFS(hash string) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	blocks, err := bc.ipfsHandler.RetrieveBlockchain(hash)
	if err != nil {
		return fmt.Errorf("failed to restore blockchain from IPFS: %v", err)
	}

	// Validate the retrieved blockchain
	for i := 1; i < len(blocks); i++ {
		currentBlock := blocks[i]
		previousBlock := blocks[i-1]

		if currentBlock.PrevHash != previousBlock.Hash {
			return fmt.Errorf("invalid blockchain data: hash mismatch at block %d", i)
		}

		if !currentBlock.ValidateBlock() {
			return fmt.Errorf("invalid blockchain data: invalid proof of work at block %d", i)
		}
	}

	bc.Blocks = blocks
	fmt.Printf("Blockchain restored from IPFS hash: %s\n", hash)
	return nil
}
