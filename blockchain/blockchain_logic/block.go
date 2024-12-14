// block.go
package blockchain_logic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
	Nonce        int64         `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
}

// CreateBlock creates a new block with the given transactions
func CreateBlock(index int64, transactions []Transaction, prevHash string, difficulty int) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Difficulty:   difficulty,
		Nonce:        0,
	}
	block.Mine()
	return block
}

// CalculateHash calculates the hash of the block
func (b *Block) CalculateHash() string {
	data, _ := json.Marshal(struct {
		Index        int64         `json:"index"`
		Timestamp    int64         `json:"timestamp"`
		Transactions []Transaction `json:"transactions"`
		PrevHash     string        `json:"prev_hash"`
		Nonce        int64         `json:"nonce"`
	}{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
		Nonce:        b.Nonce,
	})

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Mine performs the proof of work algorithm on the block
func (b *Block) Mine() {
	target := strings.Repeat("0", b.Difficulty)

	for {
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, target) {
			fmt.Printf("Block mined! Hash: %s\n", b.Hash)
			return
		}
		b.Nonce++
	}
}

// ValidateBlock validates the block's hash and proof of work
func (b *Block) ValidateBlock() bool {
	calculatedHash := b.CalculateHash()
	if calculatedHash != b.Hash {
		return false
	}

	target := strings.Repeat("0", b.Difficulty)
	return strings.HasPrefix(b.Hash, target)
}
