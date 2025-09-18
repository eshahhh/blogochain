package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
)

type MerkleTree struct {
	Root   string
	Leaves []string
}

func NewMerkleTree(transactions []string) *MerkleTree {
	tree := &MerkleTree{
		Leaves: make([]string, len(transactions)),
	}

	for i, tx := range transactions {
		hash := sha256.Sum256([]byte(tx))
		tree.Leaves[i] = hex.EncodeToString(hash[:])
	}

	tree.Root = tree.buildTree(tree.Leaves)
	return tree
}

func (mt *MerkleTree) buildTree(nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}

	if len(nodes) == 1 {
		return nodes[0]
	}

	var newLevel []string

	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	for i := 0; i < len(nodes); i += 2 {
		combined := nodes[i] + nodes[i+1]
		hash := sha256.Sum256([]byte(combined))
		newLevel = append(newLevel, hex.EncodeToString(hash[:]))
	}

	return mt.buildTree(newLevel)
}

func (mt *MerkleTree) GetRoot() string {
	return mt.Root
}
