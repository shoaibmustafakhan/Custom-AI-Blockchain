package main

import (
	"blockchain/blockchain_logic"
	"fmt"
)

func main() {
	// Initialize and train the validator
	validator := blockchain_logic.NewMLTransactionValidator()
	err := validator.Train("transactions.csv")
	if err != nil {
		fmt.Printf("Error training model: %v\n", err)
		return
	}

	// Test transactions
	testTransactions := []blockchain_logic.Transaction{
		{Sender: "Alice", Receiver: "Bob", Amount: 500},    // Should be valid
		{Sender: "Bob", Receiver: "Charlie", Amount: 1500}, // Should be invalid
		{Sender: "Alice", Receiver: "David", Amount: 300},  // Should be valid
		{Sender: "Unknown", Receiver: "Bob", Amount: 900},  // Should be suspicious
		{Sender: "Alice", Receiver: "Bob", Amount: 2000},   // Should be invalid
	}

	fmt.Println("\nTesting Transactions:")
	fmt.Println("----------------------")

	for _, tx := range testTransactions {
		valid, confidence, reason := validator.ValidateTransaction(tx)
		fmt.Printf("\nTransaction: %s -> %s (%.2f)\n", tx.Sender, tx.Receiver, tx.Amount)
		fmt.Printf("Valid: %v\n", valid)
		fmt.Printf("Confidence: %.2f%%\n", confidence*100)
		fmt.Printf("Reason: %s\n", reason)
	}
}
