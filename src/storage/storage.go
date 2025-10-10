package storage

import "vibro/src/models"

// Storage defines the data persistence interface
type Storage interface {
	// Context operations
	GetAllContexts() ([]models.Context, error)
	GetContextByID(id string) (*models.Context, error)
	CreateContext(ctx *models.Context) error
	UpdateContext(ctx *models.Context) error
	DeleteContext(id string) error

	// Item operations
	CreateItem(item *models.Item) error
	UpdateItem(item *models.Item) error
	DeleteItem(id string) error

	// Utility
	Init() error  // Initialize storage (create dirs/files)
	Close() error // Cleanup if needed
}
