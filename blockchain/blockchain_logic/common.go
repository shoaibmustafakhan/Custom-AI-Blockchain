package blockchain_logic

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// Transaction represents a single transaction in the system
type Transaction struct {
	Sender    string  `json:"sender"`
	Receiver  string  `json:"receiver"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
}

// TransactionPool manages the collection of transactions
type TransactionPool struct {
	Transactions []Transaction
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		Transactions: make([]Transaction, 0),
	}
}

// ReadTransactionsFromCSV reads and validates transactions from CSV
func ReadTransactionsFromCSV(filepath string) ([]Transaction, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	if len(header) != 3 || header[0] != "Sender" || header[1] != "Receiver" || header[2] != "Amount" {
		return nil, fmt.Errorf("invalid CSV header format")
	}

	var transactions []Transaction
	lineNum := 1

	// Read all records
	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file
		}
		lineNum++

		if len(record) != 3 {
			return nil, fmt.Errorf("invalid record at line %d", lineNum)
		}

		amount, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount at line %d: %v", lineNum, err)
		}

		tx := Transaction{
			Sender:   record[0],
			Receiver: record[1],
			Amount:   amount,
		}
		transactions = append(transactions, tx)
	}

	if len(transactions) < 5 {
		return nil, fmt.Errorf("insufficient transactions: found %d, minimum required is 5", len(transactions))
	}

	return transactions, nil
}

// PrintTransactions prints transactions in a readable format
func PrintTransactions(transactions []Transaction) {
	fmt.Println("\nTransaction List:")
	fmt.Println("------------------")
	for i, tx := range transactions {
		fmt.Printf("%d. From: %s To: %s Amount: %.2f\n",
			i+1, tx.Sender, tx.Receiver, tx.Amount)
	}
	fmt.Println("------------------")
}
