package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"vibro/src/models"
)

func TestJSONStorage(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_contexts.json")

	// Create new storage
	store, err := NewJSONStorage(filePath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Test GetAllContexts (empty)
	contexts, err := store.GetAllContexts()
	if err != nil {
		t.Fatalf("GetAllContexts failed: %v", err)
	}
	if len(contexts) != 0 {
		t.Errorf("Expected 0 contexts, got %d", len(contexts))
	}

	// Test CreateContext
	ctx := &models.Context{
		ID:        "test-id-1",
		Name:      "Test Context",
		Color:     "#FF5733",
		Icon:      "🚀",
		Items:     []models.Item{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.CreateContext(ctx)
	if err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}

	// Test GetContextByID
	retrieved, err := store.GetContextByID("test-id-1")
	if err != nil {
		t.Fatalf("GetContextByID failed: %v", err)
	}
	if retrieved.Name != "Test Context" {
		t.Errorf("Expected name 'Test Context', got '%s'", retrieved.Name)
	}

	// Test UpdateContext
	ctx.Name = "Updated Context"
	err = store.UpdateContext(ctx)
	if err != nil {
		t.Fatalf("UpdateContext failed: %v", err)
	}

	retrieved, err = store.GetContextByID("test-id-1")
	if err != nil {
		t.Fatalf("GetContextByID failed after update: %v", err)
	}
	if retrieved.Name != "Updated Context" {
		t.Errorf("Expected name 'Updated Context', got '%s'", retrieved.Name)
	}

	// Test CreateItem
	item := &models.Item{
		ID:        "item-1",
		ContextID: "test-id-1",
		Type:      models.ItemTypeText,
		Content:   "Test content",
		Position:  0,
		CreatedAt: time.Now(),
	}

	err = store.CreateItem(item)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	// Verify item was added
	retrieved, err = store.GetContextByID("test-id-1")
	if err != nil {
		t.Fatalf("GetContextByID failed after adding item: %v", err)
	}
	if len(retrieved.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(retrieved.Items))
	}
	if retrieved.Items[0].Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", retrieved.Items[0].Content)
	}

	// Test UpdateItem
	item.Content = "Updated content"
	err = store.UpdateItem(item)
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}

	retrieved, err = store.GetContextByID("test-id-1")
	if err != nil {
		t.Fatalf("GetContextByID failed after updating item: %v", err)
	}
	if retrieved.Items[0].Content != "Updated content" {
		t.Errorf("Expected content 'Updated content', got '%s'", retrieved.Items[0].Content)
	}

	// Test DeleteItem
	err = store.DeleteItem("item-1")
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	retrieved, err = store.GetContextByID("test-id-1")
	if err != nil {
		t.Fatalf("GetContextByID failed after deleting item: %v", err)
	}
	if len(retrieved.Items) != 0 {
		t.Errorf("Expected 0 items after delete, got %d", len(retrieved.Items))
	}

	// Test DeleteContext
	err = store.DeleteContext("test-id-1")
	if err != nil {
		t.Fatalf("DeleteContext failed: %v", err)
	}

	_, err = store.GetContextByID("test-id-1")
	if err == nil {
		t.Error("Expected error when getting deleted context, got nil")
	}

	// Test GetAllContexts (should be empty again)
	contexts, err = store.GetAllContexts()
	if err != nil {
		t.Fatalf("GetAllContexts failed: %v", err)
	}
	if len(contexts) != 0 {
		t.Errorf("Expected 0 contexts after delete, got %d", len(contexts))
	}
}

func TestAtomicWriteAndBackup(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_contexts.json")

	// Create storage and add a context
	store, err := NewJSONStorage(filePath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := &models.Context{
		ID:        "test-id-1",
		Name:      "Test Context",
		Color:     "#FF5733",
		Icon:      "🚀",
		Items:     []models.Item{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.CreateContext(ctx)
	if err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}

	// Verify main file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Main storage file was not created")
	}

	// Create another context to trigger backup
	ctx2 := &models.Context{
		ID:        "test-id-2",
		Name:      "Test Context 2",
		Color:     "#33FF57",
		Icon:      "💻",
		Items:     []models.Item{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.CreateContext(ctx2)
	if err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}

	// Verify backup file exists
	backupPath := filePath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	store.Close()
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_contexts.json")

	// Create storage and add data
	store1, err := NewJSONStorage(filePath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := &models.Context{
		ID:        "test-id-1",
		Name:      "Persistent Context",
		Color:     "#FF5733",
		Icon:      "🚀",
		Items:     []models.Item{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store1.CreateContext(ctx)
	if err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}
	store1.Close()

	// Create new storage instance and verify data persists
	store2, err := NewJSONStorage(filePath)
	if err != nil {
		t.Fatalf("Failed to create second storage instance: %v", err)
	}
	defer store2.Close()

	contexts, err := store2.GetAllContexts()
	if err != nil {
		t.Fatalf("GetAllContexts failed: %v", err)
	}

	if len(contexts) != 1 {
		t.Errorf("Expected 1 persisted context, got %d", len(contexts))
	}

	if len(contexts) > 0 && contexts[0].Name != "Persistent Context" {
		t.Errorf("Expected name 'Persistent Context', got '%s'", contexts[0].Name)
	}
}
