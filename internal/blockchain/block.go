package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Block struct {
	Index        int       `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	Transactions []string  `json:"transactions"`
	PrevHash     string    `json:"prev_hash"`
	Hash         string    `json:"hash"`
	Nonce        int       `json:"nonce"`
	MerkleRoot   string    `json:"merkle_root"`
}

func NewBlock(index int, transactions []string, prevHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
	}

	block.MerkleRoot = block.calculateMerkleRoot()
	block.Hash = block.calculateHash()
	fmt.Printf("Block.hash is: %s", block.Hash)
	return block
}

func (b *Block) calculateHash() string {
	data := fmt.Sprintf("%d%s%s%d%s", b.Index, b.Timestamp.String(), strings.Join(b.Transactions, ""), b.Nonce, b.PrevHash)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (b *Block) calculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return ""
	}

	merkleTree := NewMerkleTree(b.Transactions)
	return merkleTree.GetRoot()
}

func (b *Block) Mine(difficulty int) {
	target := strings.Repeat("0", difficulty)

	for {
		b.Hash = b.calculateHash()
		if strings.HasPrefix(b.Hash, target) {
			fmt.Printf("Block %d mined! Nonce: %d, Hash: %s\n", b.Index, b.Nonce, b.Hash)
			break
		}
		b.Nonce++

		if b.Nonce%10000 == 0 {
			fmt.Printf("Mining block %d... Nonce: %d, Current hash: %s\n", b.Index, b.Nonce, b.Hash)
		}
	}
}

func (b *Block) IsValid(difficulty int) bool {
	target := strings.Repeat("0", difficulty)
	return strings.HasPrefix(b.Hash, target) && b.Hash == b.calculateHash()
}

func (b *Block) ToJSON() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
