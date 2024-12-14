// ipfs_handler.go
package blockchain_logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSHandler struct {
	shell *shell.Shell
	ctx   context.Context
}

// NewIPFSHandler creates a new IPFS handler
func NewIPFSHandler(nodeAddr string) (*IPFSHandler, error) {
	sh := shell.NewShell(nodeAddr)
	ctx := context.Background()

	// Test connection
	if _, err := sh.ID(); err != nil {
		return nil, fmt.Errorf("failed to connect to IPFS node: %v", err)
	}

	return &IPFSHandler{
		shell: sh,
		ctx:   ctx,
	}, nil
}

// StoreBlock stores a block in IPFS and returns its hash
func (ih *IPFSHandler) StoreBlock(block *Block) (string, error) {
	blockData, err := json.Marshal(block)
	if err != nil {
		return "", fmt.Errorf("failed to marshal block: %v", err)
	}

	hash, err := ih.shell.Add(bytes.NewReader(blockData))
	if err != nil {
		return "", fmt.Errorf("failed to add block to IPFS: %v", err)
	}

	return hash, nil
}

// RetrieveBlock retrieves a block from IPFS using its hash
func (ih *IPFSHandler) RetrieveBlock(hash string) (*Block, error) {
	reader, err := ih.shell.Cat(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve block from IPFS: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read block data: %v", err)
	}

	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %v", err)
	}

	return &block, nil
}

// StoreBlockchain stores the entire blockchain in IPFS
func (ih *IPFSHandler) StoreBlockchain(blockchain *Blockchain) (string, error) {
	blockchainData, err := json.Marshal(blockchain.Blocks)
	if err != nil {
		return "", fmt.Errorf("failed to marshal blockchain: %v", err)
	}

	hash, err := ih.shell.Add(bytes.NewReader(blockchainData))
	if err != nil {
		return "", fmt.Errorf("failed to add blockchain to IPFS: %v", err)
	}

	return hash, nil
}

// RetrieveBlockchain retrieves the entire blockchain from IPFS
func (ih *IPFSHandler) RetrieveBlockchain(hash string) ([]*Block, error) {
	reader, err := ih.shell.Cat(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blockchain from IPFS: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read blockchain data: %v", err)
	}

	var blocks []*Block
	if err := json.Unmarshal(data, &blocks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blockchain: %v", err)
	}

	return blocks, nil
}

// Pin pins content to ensure it's kept in the IPFS network
func (ih *IPFSHandler) Pin(hash string) error {
	return ih.shell.Pin(hash)
}

// Unpin unpins content from IPFS
func (ih *IPFSHandler) Unpin(hash string) error {
	return ih.shell.Unpin(hash)
}
