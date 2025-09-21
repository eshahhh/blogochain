package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/eshahhh/blogochain/internal/blockchain"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	hashrates map[*Client]float64
	mu        sync.RWMutex

	bc *blockchain.Blockchain
}

func NewHub(bc *blockchain.Blockchain) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		hashrates:  make(map[*Client]float64),
		bc:         bc,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			log.Println("[WS] client registered")
			h.mu.Lock()
			h.clients[c] = true
			h.hashrates[c] = 0
			h.mu.Unlock()
			h.sendChainTo(c)
			h.broadcastMetrics()
		case c := <-h.unregister:
			log.Println("[WS] client unregistered")
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				delete(h.hashrates, c)
				close(c.send)
			}
			h.mu.Unlock()
			h.broadcastMetrics()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(h.clients, c)
					delete(h.hashrates, c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) StartTicker() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			h.broadcastMetrics()
		}
	}()
}

func (h *Hub) SetHashrate(c *Client, hps float64) {
	h.mu.Lock()
	h.hashrates[c] = hps
	h.mu.Unlock()
}

func (h *Hub) TotalHashrate() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var sum float64
	for _, v := range h.hashrates {
		sum += v
	}
	return sum
}

func (h *Hub) MinerCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

type outMetrics struct {
	Type           string  `json:"type"`
	Miners         int     `json:"miners"`
	TotalHashrate  float64 `json:"total_hashrate"`
	Pending        int     `json:"pending"`
	ChainLen       int     `json:"chain_len"`
	Difficulty     int     `json:"difficulty"`
	ServerHashrate float64 `json:"server_hashrate"`
}

type outChain struct {
	Type   string              `json:"type"`
	Blocks []*blockchain.Block `json:"blocks"`
}

type outResponse struct {
	Type    string              `json:"type"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Block   *blockchain.Block   `json:"block,omitempty"`
	Results []*blockchain.Block `json:"results,omitempty"`
	Data    interface{}         `json:"data,omitempty"`
}

type outPendingTransactions struct {
	Type         string   `json:"type"`
	Transactions []string `json:"transactions"`
}

type outMiningStatus struct {
	Type       string `json:"type"`
	Mining     bool   `json:"mining"`
	BlockIndex int    `json:"block_index"`
	Difficulty int    `json:"difficulty"`
}

func (h *Hub) broadcastMetrics() {
	pending := len(h.bc.GetPendingTransactions())
	chainLen := len(h.bc.GetChain())
	m := outMetrics{
		Type:           "metrics",
		Miners:         h.MinerCount(),
		TotalHashrate:  h.TotalHashrate(),
		Pending:        pending,
		ChainLen:       chainLen,
		Difficulty:     h.bc.GetDifficulty(),
		ServerHashrate: h.bc.LastHashrate(),
	}
	h.BroadcastJSON(m)
}

func (h *Hub) BroadcastChain() {
	payload := outChain{Type: "chain", Blocks: h.bc.GetChain()}
	h.BroadcastJSON(payload)
}

func (h *Hub) sendChainTo(c *Client) {
	payload := outChain{Type: "chain", Blocks: h.bc.GetChain()}
	b, _ := json.Marshal(payload)
	select {
	case c.send <- b:
	default:
	}
}

func (h *Hub) BroadcastJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("broadcast marshal error:", err)
		return
	}
	select {
	case h.broadcast <- b:
	default:
	}
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	name string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type inboundMsg struct {
	Type       string  `json:"type"`
	Name       string  `json:"name,omitempty"`
	HPS        float64 `json:"hps,omitempty"`
	Data       string  `json:"data,omitempty"`
	Difficulty *int    `json:"difficulty,omitempty"`
	Query      string  `json:"query,omitempty"`
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Upgrade failed", http.StatusBadRequest)
		return
	}
	client := &Client{hub: s.hub, conn: conn, send: make(chan []byte, 256)}
	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg inboundMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "hello":
			c.name = msg.Name
		case "hashrate":
			c.hub.SetHashrate(c, msg.HPS)
		case "add_transaction":
			c.handleAddTransaction(msg)
		case "mine_block":
			c.handleMineBlock()
		case "set_difficulty":
			c.handleSetDifficulty(msg)
		case "search_chain":
			c.handleSearchChain(msg)
		case "get_pending":
			c.handleGetPending()
		case "get_chain":
			c.handleGetChain()
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendResponse(msgType string, success bool, message string, data interface{}) {
	response := outResponse{
		Type:    msgType,
		Success: success,
		Message: message,
		Data:    data,
	}
	c.sendJSON(response)
}

func (c *Client) sendJSON(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("client marshal error:", err)
		return
	}
	select {
	case c.send <- b:
	default:
	}
}

func (c *Client) handleAddTransaction(msg inboundMsg) {
	if msg.Data == "" {
		c.sendResponse("add_transaction_response", false, "Transaction data cannot be empty", nil)
		return
	}

	c.hub.bc.AddTransaction(msg.Data)
	c.sendResponse("add_transaction_response", true, "Transaction added successfully", nil)
	log.Printf("[WS] Transaction added: %s", msg.Data)
}

func (c *Client) handleMineBlock() {
	log.Println("[WS] Mining block requested")

	chainLen := len(c.hub.bc.GetChain())
	difficulty := c.hub.bc.GetDifficulty()

	miningStatus := outMiningStatus{
		Type:       "mining_status",
		Mining:     true,
		BlockIndex: chainLen + 1,
		Difficulty: difficulty,
	}
	c.hub.BroadcastJSON(miningStatus)

	block := c.hub.bc.MineBlock()

	miningStatus.Mining = false
	c.hub.BroadcastJSON(miningStatus)

	if block == nil {
		c.sendResponse("mine_block_response", false, "No pending transactions to mine", nil)
	} else {
		c.sendResponse("mine_block_response", true, "Block mined successfully", map[string]interface{}{"block": block})
		c.hub.BroadcastChain()
		log.Printf("[WS] Block mined: #%d", block.Index)
	}
}

func (c *Client) handleSetDifficulty(msg inboundMsg) {
	if msg.Difficulty == nil {
		c.sendResponse("set_difficulty_response", false, "Difficulty value is required", nil)
		return
	}

	c.hub.bc.SetDifficulty(*msg.Difficulty)
	newDifficulty := c.hub.bc.GetDifficulty()
	c.sendResponse("set_difficulty_response", true, "Difficulty updated", map[string]interface{}{"difficulty": newDifficulty})
	log.Printf("[WS] Difficulty set to: %d", newDifficulty)
}

func (c *Client) handleSearchChain(msg inboundMsg) {
	if msg.Query == "" {
		c.sendResponse("search_chain_response", false, "Query parameter is required", nil)
		return
	}

	results := c.hub.bc.SearchData(msg.Query)
	response := outResponse{
		Type:    "search_chain_response",
		Success: true,
		Message: "Search completed",
		Results: results,
	}
	c.sendJSON(response)
	log.Printf("[WS] Search query: %s, results: %d", msg.Query, len(results))
}

func (c *Client) handleGetPending() {
	pending := c.hub.bc.GetPendingTransactions()
	response := outPendingTransactions{
		Type:         "pending_transactions",
		Transactions: pending,
	}
	c.sendJSON(response)
}

func (c *Client) handleGetChain() {
	c.hub.sendChainTo(c)
}
