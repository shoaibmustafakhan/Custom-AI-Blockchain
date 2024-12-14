package blockchain_logic

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

// MLTransactionValidator represents our ML model
type MLTransactionValidator struct {
	weights        []float64
	bias           float64
	senderCounts   map[string]int
	receiverCounts map[string]int
	meanAmount     float64
	stdAmount      float64
	maxAmount      float64
	minAmount      float64
	// New fields for pattern recognition
	senderAverages   map[string]float64
	receiverAverages map[string]float64
}

func NewMLTransactionValidator() *MLTransactionValidator {
	return &MLTransactionValidator{
		weights:          make([]float64, 5), // Increased features
		senderCounts:     make(map[string]int),
		receiverCounts:   make(map[string]int),
		senderAverages:   make(map[string]float64),
		receiverAverages: make(map[string]float64),
	}
}

func (mv *MLTransactionValidator) Train(filepath string) error {
	// Read training data
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	// Skip header
	records = records[1:]

	// First pass: collect statistics
	var sumAmount float64
	var amounts []float64
	senderTotals := make(map[string]float64)
	receiverTotals := make(map[string]float64)

	mv.maxAmount = 0
	mv.minAmount = math.MaxFloat64

	for _, record := range records {
		amount, _ := strconv.ParseFloat(record[2], 64)
		sender, receiver := record[0], record[1]

		// Update statistics
		sumAmount += amount
		amounts = append(amounts, amount)
		mv.senderCounts[sender]++
		mv.receiverCounts[receiver]++
		senderTotals[sender] += amount
		receiverTotals[receiver] += amount

		// Update min/max
		if amount > mv.maxAmount {
			mv.maxAmount = amount
		}
		if amount < mv.minAmount {
			mv.minAmount = amount
		}
	}

	// Calculate averages
	for sender, total := range senderTotals {
		mv.senderAverages[sender] = total / float64(mv.senderCounts[sender])
	}
	for receiver, total := range receiverTotals {
		mv.receiverAverages[receiver] = total / float64(mv.receiverCounts[receiver])
	}

	// Calculate mean and standard deviation
	mv.meanAmount = sumAmount / float64(len(records))
	var sumSquares float64
	for _, amount := range amounts {
		diff := amount - mv.meanAmount
		sumSquares += diff * diff
	}
	mv.stdAmount = math.Sqrt(sumSquares / float64(len(amounts)))

	fmt.Printf("\nModel Training Statistics:\n")
	fmt.Printf("Number of transactions: %d\n", len(records))
	fmt.Printf("Average amount: %.2f\n", mv.meanAmount)
	fmt.Printf("Standard deviation: %.2f\n", mv.stdAmount)
	fmt.Printf("Min amount: %.2f\n", mv.minAmount)
	fmt.Printf("Max amount: %.2f\n", mv.maxAmount)
	fmt.Printf("Unique senders: %d\n", len(mv.senderCounts))
	fmt.Printf("Unique receivers: %d\n", len(mv.receiverCounts))

	// Train the model using logistic regression
	mv.trainLogisticRegression(records)

	return nil
}

func (mv *MLTransactionValidator) trainLogisticRegression(records [][]string) {
	learningRate := 0.01
	epochs := 100

	// Initialize weights
	for i := range mv.weights {
		mv.weights[i] = 0.01
	}
	mv.bias = 0.01

	for epoch := 0; epoch < epochs; epoch++ {
		totalLoss := 0.0

		for _, record := range records {
			amount, _ := strconv.ParseFloat(record[2], 64)
			features := mv.extractFeatures(record[0], record[1], amount)

			// Define label (1 for valid, 0 for invalid)
			label := 1.0
			if amount > 1000 || amount <= 0 {
				label = 0.0
			}

			// Forward pass
			prediction := mv.predict(features)
			loss := label - prediction

			// Update weights
			for i := range mv.weights {
				mv.weights[i] += learningRate * loss * features[i]
			}
			mv.bias += learningRate * loss

			totalLoss += math.Abs(loss)
		}

		if epoch%20 == 0 {
			fmt.Printf("Epoch %d, Average Loss: %.4f\n", epoch, totalLoss/float64(len(records)))
		}
	}
}

func (mv *MLTransactionValidator) extractFeatures(sender, receiver string, amount float64) []float64 {
	// Feature 1: Normalized amount
	normalizedAmount := (amount - mv.meanAmount) / mv.stdAmount

	// Feature 2: Sender frequency (normalized)
	senderFreq := float64(mv.senderCounts[sender]) / float64(len(mv.senderCounts))

	// Feature 3: Receiver frequency (normalized)
	receiverFreq := float64(mv.receiverCounts[receiver]) / float64(len(mv.receiverCounts))

	// Feature 4: Sender's average transaction difference
	senderAvgDiff := math.Abs(amount-mv.senderAverages[sender]) / mv.maxAmount

	// Feature 5: Receiver's average transaction difference
	receiverAvgDiff := math.Abs(amount-mv.receiverAverages[receiver]) / mv.maxAmount

	return []float64{
		normalizedAmount,
		senderFreq,
		receiverFreq,
		senderAvgDiff,
		receiverAvgDiff,
	}
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func (mv *MLTransactionValidator) predict(features []float64) float64 {
	sum := mv.bias
	for i, feature := range features {
		sum += feature * mv.weights[i]
	}
	return sigmoid(sum)
}

func (mv *MLTransactionValidator) ValidateTransaction(tx Transaction) (bool, float64, string) {
	features := mv.extractFeatures(tx.Sender, tx.Receiver, tx.Amount)
	probability := mv.predict(features)

	// Decision making with explanation
	var reason string
	if probability < 0.5 {
		if tx.Amount > 1000 {
			reason = "Amount exceeds normal transaction range"
		} else if tx.Amount <= 0 {
			reason = "Invalid transaction amount"
		} else {
			reason = "Unusual transaction pattern detected"
		}
		return false, probability, reason
	}

	return true, probability, "Transaction appears valid"
}
