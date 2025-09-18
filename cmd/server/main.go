package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eshahhh/blogochain/internal/api"
	"github.com/eshahhh/blogochain/internal/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain(0)

	server := api.NewServer(bc)

	mux := server.SetupRoutes()

	port := ":8080"
	fmt.Printf("Starting blockchain server on port %s\n", port)
	fmt.Printf("Access the web interface at http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, mux))
}
