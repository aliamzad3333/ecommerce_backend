package main

import (
	"log"

	"ecommerce-backend/internal/server"
)

func main() {
	// Create and start server
	srv, err := server.New()
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Start server with graceful shutdown
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
