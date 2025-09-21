# Blogochain - Simple Blockchain

A simple blockchain implementation in Go featuring proof-of-work mining, Merkle trees, and a web interface.

## Features

- ✅ **Block Structure**: Each block contains Index, Timestamp, Transactions, PrevHash, Hash, Nonce, and MerkleRoot
- ✅ **Genesis Block**: Automatic creation of the first block in the blockchain
- ✅ **Merkle Tree**: Efficient and secure storage of transactions using Merkle trees
- ✅ **Transaction Management**: Add transactions as strings to pending pool
- ✅ **Proof of Work Mining**: Mine blocks using proof-of-work algorithm with adjustable difficulty
- ✅ **Blockchain Viewer**: View the complete blockchain through web interface
- ✅ **Search Functionality**: Search for data within the blockchain

## Prerequisites

- Go 1.21 or later
- Web browser for accessing the interface

## Installation

1. Clone the repository:
```bash
git clone https://github.com/eshahhh/blogochain.git
cd blogochain
```

2. Initialize Go modules:
```bash
go mod tidy
```

## Running the Application

1. Start the server:
```bash
go run cmd/server/main.go
```

2. Open your web browser and navigate to:
```
http://localhost:8080
```

## Usage

### Web Interface

1. **Add Transaction**: Enter transaction data in the text field and click "Add Transaction"
2. **Mine Block**: Click "Mine Block" to mine a new block with all pending transactions.
3. **View Blockchain**: The blockchain is automatically displayed and updates after mining.
4. **Search**: Enter a search term to find blocks containing specific data

## Technical Details

### Block Structure

Each block contains:
- **Index**: Block height in the chain
- **Timestamp**: When the block was created
- **Transactions**: Array of transaction strings
- **PrevHash**: Hash of the previous block
- **Hash**: Current block's hash (computed)
- **Nonce**: Proof-of-work nonce
- **MerkleRoot**: Root hash of the transaction Merkle tree
- **Difficulty**: Difficulty actually used to mine this specific block (0 for fast-mined demo blocks)

### Proof of Work

The mining algorithm uses a simple proof-of-work system:
- Difficulty is configurable (number of leading zeros required)
- Nonce is incremented until the hash meets the difficulty requirement
- Hash is computed as SHA256 of block data

### Merkle Tree

- Transactions are hashed and arranged in a binary tree
- Root hash provides tamper-proof verification of all transactions
- Handles odd numbers of transactions by duplicating the last one

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## AI usage
I have used AI for the websockets part, quality of life work (add print statements, make the cli interface using our functions, fix formatting/variable names), a little bit of debugging fixes and the skeleton for this README. Definitely less than 30% though!