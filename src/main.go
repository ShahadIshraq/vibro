package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vibro/src/server"
	"vibro/src/storage"
)

const (
	serverAddr = "0.0.0.0:8080"
	dataFile   = "./data/contexts.json"
	uploadsDir = "./uploads"
)

func main() {
	// Initialize storage
	if err := initDirectories(); err != nil {
		log.Fatalf("Failed to initialize directories: %v", err)
	}

	store, err := storage.NewJSONStorage(dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create server
	srv := server.New(serverAddr, store)

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on http://%s", serverAddr)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func initDirectories() error {
	dirs := []string{"./data", uploadsDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
