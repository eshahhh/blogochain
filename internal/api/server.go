package api

import (
	"net/http"

	"github.com/eshahhh/blogochain/internal/blockchain"
)

type Server struct {
	blockchain *blockchain.Blockchain
	hub        *Hub
}

func NewServer(bc *blockchain.Blockchain) *Server {
	s := &Server{blockchain: bc}
	h := NewHub(bc)
	s.hub = h
	go h.Run()
	h.StartTicker()
	return s
}

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", s.HandleWS)

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/static/index.html")
	})

	return mux
}
