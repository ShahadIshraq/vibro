package main

import (
	"context"
	"flag"
	"fmt"
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
	dataFile   = "./data/contexts.json"
	uploadsDir = "./uploads"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	// Construct server address
	serverAddr := fmt.Sprintf("0.0.0.0:%d", *port)

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
