package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vibro/src/models"
)

// JSONStorage implements the Storage interface using JSON file persistence
type JSONStorage struct {
	filePath string
	mu       sync.RWMutex // Thread-safe reads/writes
	contexts []models.Context
}

// NewJSONStorage creates a new JSON-based storage
func NewJSONStorage(filePath string) (*JSONStorage, error) {
	s := &JSONStorage{
		filePath: filePath,
		contexts: make([]models.Context, 0),
	}

	// Load existing data or create new file
	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

// Init initializes the storage (creates directories if needed)
func (s *JSONStorage) Init() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	return nil
}

// Close performs cleanup (no-op for file storage)
func (s *JSONStorage) Close() error {
	return nil
}

// load reads the JSON file into memory
func (s *JSONStorage) load() error {
	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// File doesn't exist, create empty file
		if err := s.Init(); err != nil {
			return err
		}
		return s.save()
	}

	// Read file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	// Parse JSON
	if len(data) == 0 {
		s.contexts = make([]models.Context, 0)
		return nil
	}

	if err := json.Unmarshal(data, &s.contexts); err != nil {
		return fmt.Errorf("failed to parse storage file: %w", err)
	}

	return nil
}

// save writes in-memory data to JSON file with atomic writes
func (s *JSONStorage) save() error {
	// Create backup of existing file
	if _, err := os.Stat(s.filePath); err == nil {
		backupPath := s.filePath + ".bak"
		if err := copyFile(s.filePath, backupPath); err != nil {
			// Log but don't fail on backup error
			fmt.Printf("Warning: failed to create backup: %v\n", err)
		}
	}

	// Marshal data to JSON
	data, err := json.MarshalIndent(s.contexts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write to temporary file first (atomic write)
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temp file to actual file (atomic on POSIX)
	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file on failure
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// GetAllContexts returns all contexts
func (s *JSONStorage) GetAllContexts() ([]models.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	contexts := make([]models.Context, len(s.contexts))
	copy(contexts, s.contexts)
	return contexts, nil
}

// GetContextByID finds a context by ID
func (s *JSONStorage) GetContextByID(id string) (*models.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.contexts {
		if s.contexts[i].ID == id {
			// Return a copy
			ctx := s.contexts[i]
			return &ctx, nil
		}
	}

	return nil, errors.New("context not found")
}

// CreateContext adds a new context and saves to disk
func (s *JSONStorage) CreateContext(ctx *models.Context) error {
	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid context: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate ID
	for i := range s.contexts {
		if s.contexts[i].ID == ctx.ID {
			return errors.New("context with this ID already exists")
		}
	}

	// Add context
	s.contexts = append(s.contexts, *ctx)

	// Save to disk
	return s.save()
}

// UpdateContext updates an existing context and saves to disk
func (s *JSONStorage) UpdateContext(ctx *models.Context) error {
	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid context: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update context
	for i := range s.contexts {
		if s.contexts[i].ID == ctx.ID {
			ctx.UpdatedAt = time.Now()
			s.contexts[i] = *ctx
			return s.save()
		}
	}

	return errors.New("context not found")
}

// DeleteContext removes a context and saves to disk
func (s *JSONStorage) DeleteContext(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove context
	for i := range s.contexts {
		if s.contexts[i].ID == id {
			s.contexts = append(s.contexts[:i], s.contexts[i+1:]...)
			return s.save()
		}
	}

	return errors.New("context not found")
}

// CreateItem adds an item to a context and saves to disk
func (s *JSONStorage) CreateItem(item *models.Item) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find context and add item
	for i := range s.contexts {
		if s.contexts[i].ID == item.ContextID {
			// Check for duplicate item ID
			for j := range s.contexts[i].Items {
				if s.contexts[i].Items[j].ID == item.ID {
					return errors.New("item with this ID already exists")
				}
			}

			s.contexts[i].Items = append(s.contexts[i].Items, *item)
			s.contexts[i].UpdatedAt = time.Now()
			return s.save()
		}
	}

	return errors.New("context not found")
}

// UpdateItem updates an existing item and saves to disk
func (s *JSONStorage) UpdateItem(item *models.Item) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find context containing the item
	for i := range s.contexts {
		for j := range s.contexts[i].Items {
			if s.contexts[i].Items[j].ID == item.ID {
				s.contexts[i].Items[j] = *item
				s.contexts[i].UpdatedAt = time.Now()
				return s.save()
			}
		}
	}

	return errors.New("item not found")
}

// DeleteItem removes an item from a context and saves to disk
func (s *JSONStorage) DeleteItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find context containing the item
	for i := range s.contexts {
		for j := range s.contexts[i].Items {
			if s.contexts[i].Items[j].ID == id {
				s.contexts[i].Items = append(s.contexts[i].Items[:j], s.contexts[i].Items[j+1:]...)
				s.contexts[i].UpdatedAt = time.Now()
				return s.save()
			}
		}
	}

	return errors.New("item not found")
}

// ReorderContexts updates the order of contexts and saves to disk
func (s *JSONStorage) ReorderContexts(contexts []models.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Replace the contexts array with the reordered one
	s.contexts = contexts

	// Save to disk
	return s.save()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
