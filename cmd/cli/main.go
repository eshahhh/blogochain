package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eshahhh/blogochain/internal/blockchain"
)

func main() {
	fmt.Println("Blogochain CLI")
	fmt.Println("Type 'help' for available commands or 'exit' to quit")
	fmt.Println(strings.Repeat("=", 50))

	if len(os.Args) > 1 && os.Args[1] != "interactive" {
		executeCommand(os.Args[1:])
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nblogochain> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		args := strings.Fields(input)
		executeCommand(args)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

func executeCommand(args []string) {
	if len(args) == 0 {
		return
	}

	command := args[0]

	oldArgs := os.Args
	os.Args = append([]string{"cli"}, args...)
	defer func() { os.Args = oldArgs }()

	switch command {
	case "create-chain":
		createChain()
	case "mine-block":
		mineBlock()
	case "add-tx":
		addTransaction()
	case "show-chain":
		showChain()
	case "validate":
		validateChain()
	case "search":
		searchTransactions()
	case "help":
		printUsage()
	case "status":
		showStatus()
	case "clear":
		clearScreen()
	case "reset":
		resetBlockchain()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type 'help' for available commands")
	}
}

func printUsage() {
	fmt.Println("Blogochain CLI")
	fmt.Println("Usage: ./cli [interactive] or ./cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-chain [difficulty]    - Create a new blockchain with specified difficulty")
	fmt.Println("  mine-block                   - Mine a block with pending transactions")
	fmt.Println("  add-tx <data>                - Add a transaction to pending pool")
	fmt.Println("  show-chain                   - Display the entire blockchain")
	fmt.Println("  validate                     - Validate the blockchain integrity")
	fmt.Println("  search <query>               - Search transactions across all blocks")
	fmt.Println("  status                       - Show blockchain status")
	fmt.Println("  clear                        - Clear the screen")
	fmt.Println("  reset                        - Reset the blockchain")
	fmt.Println("  help                         - Show this help message")
	fmt.Println("  exit/quit                    - Exit interactive mode")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./cli                        # Start interactive mode")
	fmt.Println("  ./cli create-chain 3         # Single command mode")
	fmt.Println("  In interactive mode:")
	fmt.Println("    add-tx \"Hello, blockchain!\"")
	fmt.Println("    mine-block")
	fmt.Println("    search \"Hello\"")
}

var globalBlockchain *blockchain.Blockchain

func getOrCreateBlockchain() *blockchain.Blockchain {
	if globalBlockchain == nil {
		difficulty := 4
		if len(os.Args) > 2 {
			if d, err := strconv.Atoi(os.Args[2]); err == nil && d >= 0 {
				difficulty = d
			}
		}
		fmt.Printf("Creating new blockchain with difficulty %d\n", difficulty)
		globalBlockchain = blockchain.NewBlockchain(difficulty)
	}
	return globalBlockchain
}

func createChain() {
	bc := getOrCreateBlockchain()
	fmt.Printf("Blockchain created with %d blocks\n", len(bc.GetChain()))
	fmt.Printf("Current difficulty: %d\n", bc.GetDifficulty())
	fmt.Printf("Pending transactions: %d\n", len(bc.GetPendingTransactions()))
}

func mineBlock() {
	bc := getOrCreateBlockchain()

	pending := bc.GetPendingTransactions()
	if len(pending) == 0 {
		fmt.Println("No pending transactions to mine")
		fmt.Println("Use 'add-tx <data>' to add transactions first")
		return
	}

	fmt.Printf("Mining block with %d pending transactions...\n", len(pending))
	fmt.Printf("Current difficulty: %d\n", bc.GetDifficulty())
	fmt.Println(strings.Repeat("-", 40))

	start := time.Now()
	block := bc.MineBlock()
	duration := time.Since(start)

	if block != nil {
		hashrate := bc.LastHashrate()
		fmt.Printf("\nBlock mined successfully!\n")
		fmt.Printf("Block #%d\n", block.Index)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("Mining time: %v\n", duration)
		fmt.Printf("Hashrate: %.2f H/s\n", hashrate)
	} else {
		fmt.Println("Mining failed")
	}
}

func addTransaction() {
	if len(os.Args) < 3 {
		fmt.Println("Please provide transaction data")
		fmt.Println("Usage: ./cli add-tx \"transaction data\"")
		return
	}

	data := strings.Join(os.Args[2:], " ")
	bc := getOrCreateBlockchain()
	bc.AddTransaction(data)

	fmt.Printf("Transaction added: %s\n", data)
	fmt.Printf("Total pending: %d\n", len(bc.GetPendingTransactions()))
}

func showChain() {
	bc := getOrCreateBlockchain()
	chain := bc.GetChain()

	fmt.Printf("Blockchain Overview\n")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total blocks: %d\n", len(chain))
	fmt.Printf("Current difficulty: %d\n", bc.GetDifficulty())
	fmt.Printf("Pending transactions: %d\n", len(bc.GetPendingTransactions()))
	fmt.Printf("Chain valid: %t\n", bc.IsValid())
	fmt.Println()

	for i, block := range chain {
		fmt.Printf("Block #%d\n", block.Index)
		fmt.Printf("  Timestamp: %s\n", block.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Hash: %s\n", block.Hash)
		fmt.Printf("  Previous: %s\n", block.PrevHash)
		fmt.Printf("  Nonce: %d\n", block.Nonce)
		fmt.Printf("  Difficulty: %d\n", block.Difficulty)
		fmt.Printf("  Merkle Root: %s\n", block.MerkleRoot)
		fmt.Printf("  Transactions (%d):\n", len(block.Transactions))
		for j, tx := range block.Transactions {
			fmt.Printf("    %d. %s\n", j+1, tx)
		}

		if i < len(chain)-1 {
			fmt.Println(strings.Repeat("-", 30))
		}
	}
}

func validateChain() {
	bc := getOrCreateBlockchain()

	fmt.Println("Validating blockchain...")
	fmt.Println(strings.Repeat("=", 30))

	start := time.Now()
	isValid := bc.IsValid()
	duration := time.Since(start)

	if isValid {
		fmt.Printf("Blockchain is valid!\n")
	} else {
		fmt.Printf("Blockchain validation failed!\n")
	}

	fmt.Printf("Validation time: %v\n", duration)
	fmt.Printf("Total blocks validated: %d\n", len(bc.GetChain()))

	chain := bc.GetChain()
	fmt.Println("\nDetailed validation:")

	for i, block := range chain {
		fmt.Printf("  Block #%d: ", block.Index)

		if i > 0 && block.PrevHash != chain[i-1].Hash {
			fmt.Printf("Previous hash mismatch\n")
		} else if len(block.Hash) != 64 {
			fmt.Printf("Invalid hash format\n")
		} else {
			fmt.Printf("OK\n")
		}
	}
}

func showStatus() {
	if globalBlockchain == nil {
		fmt.Println("Blockchain Status: Not initialized")
		fmt.Println("Use 'create-chain [difficulty]' to initialize")
		return
	}

	bc := globalBlockchain
	chain := bc.GetChain()
	pending := bc.GetPendingTransactions()

	fmt.Println("Blockchain Status")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("Total blocks: %d\n", len(chain))
	fmt.Printf("Difficulty: %d\n", bc.GetDifficulty())
	fmt.Printf("Pending transactions: %d\n", len(pending))
	fmt.Printf("Chain valid: %t\n", bc.IsValid())

	if len(chain) > 0 {
		latestBlock := bc.GetLatestBlock()
		fmt.Printf("Latest block: #%d\n", latestBlock.Index)
		fmt.Printf("Latest hash: %s\n", latestBlock.Hash[:16]+"...")
		fmt.Printf("Latest timestamp: %s\n", latestBlock.Timestamp.Format("15:04:05"))
	}

	hashrate := bc.LastHashrate()
	if hashrate > 0 {
		fmt.Printf("Last hashrate: %.2f H/s\n", hashrate)
	}
}

func searchTransactions() {
	if len(os.Args) < 3 {
		fmt.Println("Please provide search query")
		fmt.Println("Usage: ./cli search \"query\"")
		return
	}

	query := strings.Join(os.Args[2:], " ")
	bc := getOrCreateBlockchain()

	results := bc.SearchData(query)

	fmt.Printf("Search results for \"%s\":\n", query)
	fmt.Println(strings.Repeat("=", 40))

	if len(results) == 0 {
		fmt.Println("No transactions found matching the query")
		return
	}

	fmt.Printf("Found %d blocks containing the query:\n", len(results))
	fmt.Println()

	for _, block := range results {
		fmt.Printf("Block #%d (Hash: %s)\n", block.Index, block.Hash[:16]+"...")
		fmt.Printf("  Timestamp: %s\n", block.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Matching transactions:\n")

		for i, tx := range block.Transactions {
			if containsQuery(tx, query) {
				fmt.Printf("    %d. %s\n", i+1, tx)
			}
		}
		fmt.Println()
	}
}

func containsQuery(text, query string) bool {
	text = strings.ToLower(text)
	query = strings.ToLower(query)
	return strings.Contains(text, query)
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
	fmt.Println("Blogochain Interactive CLI")
	fmt.Println("Screen cleared!")
}

func resetBlockchain() {
	fmt.Print("Are you sure you want to reset the blockchain? (y/N): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response == "y" || response == "yes" {
			globalBlockchain = nil
			fmt.Println("Blockchain reset successfully!")
			fmt.Println("Use 'create-chain [difficulty]' to create a new blockchain")
		} else {
			fmt.Println("Reset cancelled")
		}
	}
}
