package api

import (
	"encoding/json"
	"net/http"

	"github.com/eshahhh/blogochain/internal/blockchain"
)

type Server struct {
	blockchain *blockchain.Blockchain
}

func NewServer(bc *blockchain.Blockchain) *Server {
	return &Server{
		blockchain: bc,
	}
}

type AddTransactionRequest struct {
	Data string `json:"data"`
}

type AddTransactionResponse struct {
	Message string `json:"message"`
}

type MineResponse struct {
	Message string            `json:"message"`
	Block   *blockchain.Block `json:"block,omitempty"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []*blockchain.Block `json:"results"`
}

func (s *Server) HandleAddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Data == "" {
		http.Error(w, "Transaction data cannot be empty", http.StatusBadRequest)
		return
	}

	s.blockchain.AddTransaction(req.Data)

	response := AddTransactionResponse{
		Message: "Transaction added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleGetPendingTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pending := s.blockchain.GetPendingTransactions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}

func (s *Server) HandleMineBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	block := s.blockchain.MineBlock()

	response := MineResponse{
		Message: "Block mined successfully",
		Block:   block,
	}

	if block == nil {
		response.Message = "No pending transactions to mine"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleGetChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chain := s.blockchain.GetChain()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chain)
}

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results := s.blockchain.SearchData(query)

	response := SearchResponse{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/tx", s.HandleAddTransaction)
	mux.HandleFunc("/api/pending", s.HandleGetPendingTransactions)
	mux.HandleFunc("/api/mine", s.HandleMineBlock)
	mux.HandleFunc("/api/chain", s.HandleGetChain)
	mux.HandleFunc("/api/search", s.HandleSearch)

	return mux
}
