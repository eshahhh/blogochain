package blockchain

import (
	"fmt"
	"sync"
)

type Blockchain struct {
	Chain      []*Block
	PendingTxs []string
	Difficulty int
	mutex      sync.RWMutex
}

func NewBlockchain(difficulty int) *Blockchain {
	bc := &Blockchain{
		Chain:      make([]*Block, 0),
		PendingTxs: make([]string, 0),
		Difficulty: difficulty,
	}

	genesisBlock := bc.createGenesisBlock()
	bc.Chain = append(bc.Chain, genesisBlock)
	fmt.Println("Blockchain created with genesis block")
	fmt.Println("Difficulty:", bc.Difficulty)
	return bc
}

func (bc *Blockchain) createGenesisBlock() *Block {
	genesisTx := []string{"Genesis Transaction - Blockchain Created"}
	block := NewBlock(0, genesisTx, "0")
	block.Mine(bc.Difficulty)
	return block
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Chain) == 0 {
		return nil
	}
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) AddTransaction(tx string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.PendingTxs = append(bc.PendingTxs, tx)
	fmt.Printf("Transaction added to pending pool: %s (Total pending: %d)\n", tx, len(bc.PendingTxs))
}

func (bc *Blockchain) MineBlock() *Block {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.PendingTxs) == 0 {
		fmt.Println("No pending transactions to mine")
		return nil
	}

	fmt.Printf("Mining new block with %d pending transactions\n", len(bc.PendingTxs))
	latestBlock := bc.GetLatestBlock()
	newBlock := NewBlock(latestBlock.Index+1, bc.PendingTxs, latestBlock.Hash)
	newBlock.Mine(bc.Difficulty)

	bc.Chain = append(bc.Chain, newBlock)
	bc.PendingTxs = []string{}
	fmt.Printf("Block %d added to chain. Pending transactions cleared.\n", newBlock.Index)

	return newBlock
}

func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		prevBlock := bc.Chain[i-1]

		if currentBlock.Hash != currentBlock.calculateHash() {
			return false
		}

		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}

		if !currentBlock.IsValid(bc.Difficulty) {
			return false
		}
	}

	return true
}

func (bc *Blockchain) GetChain() []*Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	chain := make([]*Block, len(bc.Chain))
	copy(chain, bc.Chain)
	return chain
}

func (bc *Blockchain) GetPendingTransactions() []string {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	pending := make([]string, len(bc.PendingTxs))
	copy(pending, bc.PendingTxs)
	return pending
}

func (bc *Blockchain) SearchData(query string) []*Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	var results []*Block
	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			if containsString(tx, query) {
				results = append(results, block)
				break
			}
		}
	}
	return results
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsIgnoreCase(str, substr))
}

func containsIgnoreCase(str, substr string) bool {
	str = toLower(str)
	substr = toLower(substr)
	return len(str) >= len(substr) && findSubstring(str, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
