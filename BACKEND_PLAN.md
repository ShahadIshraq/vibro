# Vibro Backend - Detailed Implementation Plan

## Overview
A lightweight, fast Go backend using only the standard library. Focus on simplicity, performance, and minimal dependencies.

---

## 1. Project Structure

### Repository Structure (What's Committed to Git)

```
vibro/
├── src/                 # All Go source code
│   ├── main.go          # Entry point, server initialization
│   ├── server/
│   │   ├── server.go    # HTTP server setup and configuration
│   │   ├── router.go    # Route registration and middleware
│   │   └── middleware.go # Logging, CORS, error recovery
│   ├── handlers/
│   │   ├── context.go   # Context CRUD handlers
│   │   ├── item.go      # Item CRUD handlers
│   │   └── upload.go    # Image upload handler
│   ├── storage/
│   │   ├── storage.go   # Storage interface and JSON implementation
│   │   └── file.go      # File system operations
│   ├── models/
│   │   └── models.go    # Data structures (Context, Item)
│   └── utils/
│       ├── response.go  # JSON response helpers
│       └── validation.go # Input validation utilities
├── frontend/            # Static files served by Go
│   ├── index.html
│   ├── style.css
│   └── app.js
├── go.mod               # Go module file (in root)
├── go.sum               # Go dependencies checksum
└── .gitignore           # Ignore data/ and uploads/
```

### Runtime Directory Structure (Created Automatically)

When you run the binary, these directories are created automatically in the **working directory** where the binary executes:

```
[working-directory]/
├── vibro                # The compiled binary
├── data/                # Auto-created at runtime (git-ignored)
│   └── contexts.json    # JSON data storage file
└── uploads/             # Auto-created at runtime (git-ignored)
    └── [images]         # User-uploaded image files
```

**Key Points:**
- All Go code lives in `src/` directory in the repository
- `go.mod` stays at project root for Go tooling compatibility
- Runtime directories (`data/`, `uploads/`) are **NOT** part of the repository
- Runtime directories are created automatically by `initDirectories()` in [main.go:1160-1170](BACKEND_PLAN.md#L1160-L1170)
- Build from root: `go build -o vibro ./src`
- Run from root (development): `./vibro` (creates `./data/` and `./uploads/` in current directory)
- Run from anywhere (production): The binary creates runtime directories relative to its working directory

---

## 2. Data Models (`models/models.go`)

### Core Structures

```go
package models

import "time"

// Context represents a user context/workspace
type Context struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Color     string    `json:"color"`     // Hex color code (e.g., "#FF5733")
    Icon      string    `json:"icon"`      // Emoji or icon identifier
    Items     []Item    `json:"items"`     // Nested items
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// Item represents content within a context
type Item struct {
    ID        string    `json:"id"`
    ContextID string    `json:"contextId"`
    Type      ItemType  `json:"type"`      // "text", "bullet", "image"
    Content   string    `json:"content"`   // Text or image path
    Position  int       `json:"position"`  // For ordering (0-indexed)
    CreatedAt time.Time `json:"createdAt"`
}

// ItemType defines the type of content
type ItemType string

const (
    ItemTypeText   ItemType = "text"
    ItemTypeBullet ItemType = "bullet"
    ItemTypeImage  ItemType = "image"
)

// Validation methods
func (c *Context) Validate() error
func (i *Item) Validate() error
```

### Request/Response DTOs

```go
// CreateContextRequest for POST /api/contexts
type CreateContextRequest struct {
    Name  string `json:"name"`
    Color string `json:"color"`
    Icon  string `json:"icon"`
}

// UpdateContextRequest for PUT /api/contexts/:id
type UpdateContextRequest struct {
    Name  string `json:"name"`
    Color string `json:"color"`
    Icon  string `json:"icon"`
}

// CreateItemRequest for POST /api/contexts/:id/items
type CreateItemRequest struct {
    Type    ItemType `json:"type"`
    Content string   `json:"content"`
}

// UpdateItemRequest for PUT /api/items/:id
type UpdateItemRequest struct {
    Content  string `json:"content"`
    Position *int   `json:"position,omitempty"` // Optional, for reordering
}

// ErrorResponse for consistent error handling
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

// SuccessResponse for operations without data
type SuccessResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
```

---

## 3. Storage Layer (`storage/`)

### Interface Design (`storage/storage.go`)

```go
package storage

import "vibro/models"

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
```

### JSON Implementation (`storage/file.go`)

```go
package storage

import (
    "encoding/json"
    "os"
    "sync"
)

type JSONStorage struct {
    filePath string
    mu       sync.RWMutex  // Thread-safe reads/writes
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

// Key methods:
// - load() - Read JSON file into memory
// - save() - Write in-memory data to JSON file
// - GetAllContexts() - Return all contexts
// - GetContextByID() - Find context by ID
// - CreateContext() - Add new context and save
// - UpdateContext() - Update existing context and save
// - DeleteContext() - Remove context and save
// - CreateItem() - Add item to context and save
// - UpdateItem() - Update item and save
// - DeleteItem() - Remove item from context and save
```

**Implementation Details:**

1. **In-Memory Cache**: Load entire JSON into memory for fast reads
2. **Write-Through**: Every mutation saves to disk immediately
3. **Thread Safety**: Use `sync.RWMutex` for concurrent access
4. **Atomic Writes**: Write to temp file, then rename (atomic on POSIX)
5. **Auto-backup**: Keep `.bak` file before writes

**Performance Optimizations:**
- Read operations: O(n) scan (acceptable for 3-10 contexts)
- Write operations: Full file rewrite (acceptable for small datasets)
- No caching complexity needed for single-user app

---

## 4. HTTP Server (`server/`)

### Server Setup (`server/server.go`)

```go
package server

import (
    "context"
    "net/http"
    "time"
)

type Server struct {
    httpServer *http.Server
    storage    storage.Storage
}

func New(addr string, storage storage.Storage) *Server {
    s := &Server{storage: storage}

    router := s.setupRouter()

    s.httpServer = &http.Server{
        Addr:         addr,
        Handler:      router,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return s
}

func (s *Server) Start() error {
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}
```

### Router (`server/router.go`)

```go
package server

import (
    "net/http"
    "vibro/handlers"
)

func (s *Server) setupRouter() http.Handler {
    mux := http.NewServeMux()

    // API routes (most specific, handled first)
    mux.HandleFunc("GET /api/contexts", handlers.GetContexts(s.storage))
    mux.HandleFunc("POST /api/contexts", handlers.CreateContext(s.storage))
    mux.HandleFunc("GET /api/contexts/{id}", handlers.GetContext(s.storage))
    mux.HandleFunc("PUT /api/contexts/{id}", handlers.UpdateContext(s.storage))
    mux.HandleFunc("DELETE /api/contexts/{id}", handlers.DeleteContext(s.storage))

    mux.HandleFunc("POST /api/contexts/{id}/items", handlers.CreateItem(s.storage))
    mux.HandleFunc("PUT /api/items/{id}", handlers.UpdateItem(s.storage))
    mux.HandleFunc("DELETE /api/items/{id}", handlers.DeleteItem(s.storage))

    mux.HandleFunc("POST /api/upload", handlers.UploadImage())

    // Serve uploaded images
    uploadsFS := http.FileServer(http.Dir("./uploads"))
    mux.Handle("/uploads/", http.StripPrefix("/uploads/", uploadsFS))

    // Serve frontend static files (catch-all, must be last)
    frontendFS := http.FileServer(http.Dir("./frontend"))
    mux.Handle("/", frontendFS)

    // Wrap with middleware
    handler := loggingMiddleware(mux)
    handler = recoveryMiddleware(handler)
    handler = corsMiddleware(handler)

    return handler
}
```

**Route Pattern Notes (Go 1.22+):**
- Use new pattern syntax: `GET /api/contexts`
- Path parameters: `{id}` extracted with `r.PathValue("id")`
- Automatic method matching (no manual method checks)

### Middleware (`server/middleware.go`)

```go
package server

import (
    "log"
    "net/http"
    "time"
)

// loggingMiddleware logs each request
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap ResponseWriter to capture status code
        lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(lw, r)

        log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.statusCode, time.Since(start))
    })
}

// recoveryMiddleware catches panics
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// corsMiddleware adds CORS headers (for local development)
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
    lw.statusCode = code
    lw.ResponseWriter.WriteHeader(code)
}
```

---

## 5. HTTP Handlers (`handlers/`)

### Context Handlers (`handlers/context.go`)

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "vibro/models"
    "vibro/storage"
    "vibro/utils"

    "github.com/google/uuid"
)

// GetContexts returns all contexts
func GetContexts(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        contexts, err := store.GetAllContexts()
        if err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch contexts", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, contexts)
    }
}

// CreateContext creates a new context
func CreateContext(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req models.CreateContextRequest

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err)
            return
        }

        // Validate input
        if err := utils.ValidateCreateContext(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, err.Error(), nil)
            return
        }

        // Create context
        ctx := &models.Context{
            ID:        uuid.New().String(),
            Name:      req.Name,
            Color:     req.Color,
            Icon:      req.Icon,
            Items:     make([]models.Item, 0),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        if err := store.CreateContext(ctx); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to create context", err)
            return
        }

        utils.RespondJSON(w, http.StatusCreated, ctx)
    }
}

// GetContext returns a specific context by ID
func GetContext(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")

        ctx, err := store.GetContextByID(id)
        if err != nil {
            utils.RespondError(w, http.StatusNotFound, "Context not found", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, ctx)
    }
}

// UpdateContext updates an existing context
func UpdateContext(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")

        var req models.UpdateContextRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err)
            return
        }

        // Get existing context
        ctx, err := store.GetContextByID(id)
        if err != nil {
            utils.RespondError(w, http.StatusNotFound, "Context not found", err)
            return
        }

        // Update fields
        ctx.Name = req.Name
        ctx.Color = req.Color
        ctx.Icon = req.Icon
        ctx.UpdatedAt = time.Now()

        if err := store.UpdateContext(ctx); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to update context", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, ctx)
    }
}

// DeleteContext deletes a context
func DeleteContext(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")

        if err := store.DeleteContext(id); err != nil {
            utils.RespondError(w, http.StatusNotFound, "Context not found", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, models.SuccessResponse{
            Success: true,
            Message: "Context deleted successfully",
        })
    }
}
```

### Item Handlers (`handlers/item.go`)

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "vibro/models"
    "vibro/storage"
    "vibro/utils"

    "github.com/google/uuid"
)

// CreateItem adds an item to a context
func CreateItem(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        contextID := r.PathValue("id")

        var req models.CreateItemRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err)
            return
        }

        // Validate
        if err := utils.ValidateCreateItem(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, err.Error(), nil)
            return
        }

        // Verify context exists
        ctx, err := store.GetContextByID(contextID)
        if err != nil {
            utils.RespondError(w, http.StatusNotFound, "Context not found", err)
            return
        }

        // Create item
        item := &models.Item{
            ID:        uuid.New().String(),
            ContextID: contextID,
            Type:      req.Type,
            Content:   req.Content,
            Position:  len(ctx.Items), // Append to end
            CreatedAt: time.Now(),
        }

        if err := store.CreateItem(item); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to create item", err)
            return
        }

        utils.RespondJSON(w, http.StatusCreated, item)
    }
}

// UpdateItem updates an existing item
func UpdateItem(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")

        var req models.UpdateItemRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err)
            return
        }

        // Get all contexts to find the item
        contexts, err := store.GetAllContexts()
        if err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch contexts", err)
            return
        }

        var item *models.Item
        for _, ctx := range contexts {
            for i := range ctx.Items {
                if ctx.Items[i].ID == id {
                    item = &ctx.Items[i]
                    break
                }
            }
            if item != nil {
                break
            }
        }

        if item == nil {
            utils.RespondError(w, http.StatusNotFound, "Item not found", nil)
            return
        }

        // Update fields
        item.Content = req.Content
        if req.Position != nil {
            item.Position = *req.Position
        }

        if err := store.UpdateItem(item); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to update item", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, item)
    }
}

// DeleteItem removes an item
func DeleteItem(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")

        if err := store.DeleteItem(id); err != nil {
            utils.RespondError(w, http.StatusNotFound, "Item not found", err)
            return
        }

        utils.RespondJSON(w, http.StatusOK, models.SuccessResponse{
            Success: true,
            Message: "Item deleted successfully",
        })
    }
}
```

### Upload Handler (`handlers/upload.go`)

```go
package handlers

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "vibro/utils"
)

const (
    maxUploadSize = 10 << 20 // 10 MB
    uploadsDir    = "./uploads"
)

type UploadResponse struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
}

// UploadImage handles image uploads
func UploadImage() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Limit upload size
        r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

        // Parse multipart form
        if err := r.ParseMultipartForm(maxUploadSize); err != nil {
            utils.RespondError(w, http.StatusBadRequest, "File too large", err)
            return
        }

        // Get file from form
        file, header, err := r.FormFile("image")
        if err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid file upload", err)
            return
        }
        defer file.Close()

        // Validate file type
        if !isValidImageType(header.Filename) {
            utils.RespondError(w, http.StatusBadRequest, "Invalid file type. Only PNG, JPG, JPEG, GIF allowed", nil)
            return
        }

        // Generate unique filename
        filename := generateFilename(header.Filename)
        filepath := filepath.Join(uploadsDir, filename)

        // Create uploads directory if needed
        if err := os.MkdirAll(uploadsDir, 0755); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to create upload directory", err)
            return
        }

        // Save file
        dst, err := os.Create(filepath)
        if err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to save file", err)
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to write file", err)
            return
        }

        // Return file URL
        response := UploadResponse{
            URL:      "/uploads/" + filename,
            Filename: filename,
        }

        utils.RespondJSON(w, http.StatusCreated, response)
    }
}

func isValidImageType(filename string) bool {
    ext := filepath.Ext(filename)
    validExts := map[string]bool{
        ".png":  true,
        ".jpg":  true,
        ".jpeg": true,
        ".gif":  true,
        ".webp": true,
    }
    return validExts[ext]
}

func generateFilename(original string) string {
    ext := filepath.Ext(original)
    hash := sha256.New()
    hash.Write([]byte(fmt.Sprintf("%s-%d", original, time.Now().UnixNano())))
    return hex.EncodeToString(hash.Sum(nil))[:16] + ext
}
```

---

## 6. Static File Serving

### Frontend File Server Configuration

The backend serves static frontend files from the `./frontend` directory. All non-API routes are handled by the file server.

**Route Priority:**
1. API routes (`/api/*`) - handled by API handlers
2. Uploaded images (`/uploads/*`) - served from uploads directory
3. Everything else (`/`) - served from frontend directory

**Updated router.go implementation:**

```go
func (s *Server) setupRouter() http.Handler {
    mux := http.NewServeMux()

    // API routes (handled first, most specific)
    mux.HandleFunc("GET /api/contexts", handlers.GetContexts(s.storage))
    mux.HandleFunc("POST /api/contexts", handlers.CreateContext(s.storage))
    mux.HandleFunc("GET /api/contexts/{id}", handlers.GetContext(s.storage))
    mux.HandleFunc("PUT /api/contexts/{id}", handlers.UpdateContext(s.storage))
    mux.HandleFunc("DELETE /api/contexts/{id}", handlers.DeleteContext(s.storage))

    mux.HandleFunc("POST /api/contexts/{id}/items", handlers.CreateItem(s.storage))
    mux.HandleFunc("PUT /api/items/{id}", handlers.UpdateItem(s.storage))
    mux.HandleFunc("DELETE /api/items/{id}", handlers.DeleteItem(s.storage))

    mux.HandleFunc("POST /api/upload", handlers.UploadImage())

    // Serve uploaded images from /uploads directory
    uploadsFS := http.FileServer(http.Dir("./uploads"))
    mux.Handle("/uploads/", http.StripPrefix("/uploads/", uploadsFS))

    // Serve frontend static files (catch-all, must be last)
    frontendFS := http.FileServer(http.Dir("./frontend"))
    mux.Handle("/", frontendFS)

    // Wrap with middleware
    handler := loggingMiddleware(mux)
    handler = recoveryMiddleware(handler)
    handler = corsMiddleware(handler)

    return handler
}
```

### Initial Frontend Files (Hello World)

The backend expects these files in the `frontend/` directory to start:

**frontend/index.html:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vibro - Context Switcher</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div id="app">
        <h1>Hello Vibro!</h1>
        <p>Context switching application</p>
        <div id="status"></div>
    </div>
    <script src="/app.js"></script>
</body>
</html>
```

**frontend/style.css:**
```css
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #f5f5f5;
    color: #333;
    padding: 20px;
}

#app {
    max-width: 1200px;
    margin: 0 auto;
    background: white;
    padding: 40px;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

h1 {
    color: #2c3e50;
    margin-bottom: 10px;
}

p {
    color: #666;
    margin-bottom: 20px;
}

#status {
    padding: 10px;
    border-radius: 4px;
    font-weight: 500;
}

#status.success {
    background: #d4edda;
    color: #155724;
}

#status.error {
    background: #f8d7da;
    color: #721c24;
}
```

**frontend/app.js:**
```javascript
// Simple API health check on page load
document.addEventListener('DOMContentLoaded', async () => {
    const statusDiv = document.getElementById('status');

    try {
        const response = await fetch('/api/contexts');

        if (response.ok) {
            const contexts = await response.json();
            statusDiv.textContent = `✓ Backend connected! Found ${contexts.length} context(s).`;
            statusDiv.className = 'success';
        } else {
            throw new Error(`HTTP ${response.status}`);
        }
    } catch (error) {
        statusDiv.textContent = `✗ Backend connection failed: ${error.message}`;
        statusDiv.className = 'error';
    }
});

// Utility function for API calls (to be used by frontend implementation)
async function apiCall(endpoint, options = {}) {
    const response = await fetch(`/api${endpoint}`, {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
        ...options,
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'API request failed');
    }

    return response.json();
}

// Export for use in frontend development
window.api = {
    call: apiCall,
};

console.log('Vibro frontend initialized');
```

### Directory Initialization

Update `main.go` to ensure all required directories exist at startup:

```go
func initDirectories() error {
    dirs := []string{
        "./data",      // JSON storage
        "./uploads",   // User-uploaded images
        "./frontend",  // Static frontend files (should exist in repo)
    }

    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }

    // Verify frontend files exist
    requiredFiles := []string{
        "./frontend/index.html",
        "./frontend/style.css",
        "./frontend/app.js",
    }

    for _, file := range requiredFiles {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            log.Printf("Warning: Required frontend file missing: %s", file)
        }
    }

    return nil
}
```

### Testing Static File Serving

**Manual Test:**
```bash
# Start the server
./vibro

# In another terminal, test file serving
curl http://localhost:8080/
# Should return index.html content

curl http://localhost:8080/style.css
# Should return CSS content

curl http://localhost:8080/app.js
# Should return JavaScript content

# Test API endpoint
curl http://localhost:8080/api/contexts
# Should return JSON array (empty or with contexts)
```

**Browser Test:**
1. Start server: `./vibro`
2. Open browser: `http://localhost:8080`
3. Should see "Hello Vibro!" page
4. Status message should show green "✓ Backend connected!" if API is working
5. If backend isn't fully implemented, will show red error (expected)

**Expected Behavior:**
- Root path `/` → serves `index.html`
- `/style.css` → serves `style.css`
- `/app.js` → serves `app.js`
- `/api/contexts` → returns JSON from API handler
- `/uploads/image.png` → serves uploaded image (if exists)
- Any other path → 404 from frontend file server

---

## 7. Utilities (`utils/`)

### Response Helpers (`utils/response.go`)

```go
package utils

import (
    "encoding/json"
    "log"
    "net/http"
    "vibro/models"
)

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Error encoding JSON: %v", err)
    }
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, message string, err error) {
    if err != nil {
        log.Printf("Error: %s - %v", message, err)
    }

    response := models.ErrorResponse{
        Error:   http.StatusText(status),
        Message: message,
    }

    RespondJSON(w, status, response)
}
```

### Validation (`utils/validation.go`)

```go
package utils

import (
    "errors"
    "strings"
    "vibro/models"
)

func ValidateCreateContext(req *models.CreateContextRequest) error {
    if strings.TrimSpace(req.Name) == "" {
        return errors.New("context name is required")
    }

    if len(req.Name) > 100 {
        return errors.New("context name too long (max 100 characters)")
    }

    if req.Color != "" && !isValidHexColor(req.Color) {
        return errors.New("invalid color format (use hex: #RRGGBB)")
    }

    return nil
}

func ValidateCreateItem(req *models.CreateItemRequest) error {
    if req.Type != models.ItemTypeText &&
       req.Type != models.ItemTypeBullet &&
       req.Type != models.ItemTypeImage {
        return errors.New("invalid item type")
    }

    if strings.TrimSpace(req.Content) == "" {
        return errors.New("item content is required")
    }

    return nil
}

func isValidHexColor(color string) bool {
    if len(color) != 7 || color[0] != '#' {
        return false
    }

    for _, c := range color[1:] {
        if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
            return false
        }
    }

    return true
}
```

---

## 7. Main Entry Point (`main.go`)

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    "vibro/server"
    "vibro/storage"
)

const (
    serverAddr   = "localhost:8080"
    dataFile     = "./data/contexts.json"
    uploadsDir   = "./uploads"
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
```

---

## 8. Performance Optimizations

### Memory Management
- **In-memory storage**: Entire dataset loaded once, fast reads
- **Small footprint**: 10 contexts × 100 items × 1KB = ~1MB
- **No caching needed**: Direct memory access is fast enough

### Concurrency
- **RWMutex**: Multiple concurrent reads, exclusive writes
- **Goroutine-safe**: All handlers can run concurrently
- **No connection pooling needed**: SQLite/DB complexity avoided

### File I/O
- **Atomic writes**: Temp file + rename (prevents corruption)
- **Debouncing**: Frontend handles auto-save throttling
- **Backup on write**: Keep last known good state

### HTTP Server
- **Timeouts**: Prevent resource exhaustion
- **MaxBytesReader**: Limit upload sizes
- **Keep-Alive**: HTTP/1.1 persistent connections

---

## 9. Error Handling Strategy

### Levels
1. **Validation errors**: 400 Bad Request (client fault)
2. **Not found errors**: 404 Not Found (resource doesn't exist)
3. **Storage errors**: 500 Internal Server Error (server fault)
4. **Panics**: Recovered by middleware, logged, 500 response

### Logging
- All errors logged with context
- Request/response logging via middleware
- Panic stack traces captured

### Client-Friendly Responses
```json
{
  "error": "Bad Request",
  "message": "Context name is required"
}
```

---

## 10. Testing Strategy

### Unit Tests
- `models/`: Validation logic
- `utils/`: Response and validation helpers
- `storage/`: CRUD operations with temp files

### Integration Tests
- `handlers/`: Full request/response cycle with mock storage
- Test all API endpoints
- Test error scenarios

### Test Structure
```go
func TestCreateContext(t *testing.T) {
    store := &mockStorage{}
    handler := handlers.CreateContext(store)

    // Test cases...
}
```

---

## 11. Build and Deployment

### Build Command
```bash
# Build from project root (where go.mod is located)
go build -o vibro ./src
```

**Important:**
- Run the build command from the project root directory
- The `./src` tells Go to build the package in the src directory
- The binary `vibro` will be created in the project root
- Run with: `./vibro` (from project root)

### Cross-Compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o vibro-linux ./src

# Windows
GOOS=windows GOARCH=amd64 go build -o vibro.exe ./src

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o vibro-macos-intel ./src

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o vibro-macos-arm ./src
```

### Distribution

When distributing the application, package only what's needed to run (not the source code):

```
vibro-release/
├── vibro (or vibro.exe)     # Compiled binary
├── frontend/                # Static frontend files (must be included)
│   ├── index.html
│   ├── style.css
│   └── app.js
└── README.txt               # Instructions for users
```

**Distribution Notes:**
- Include only the compiled binary and `frontend/` folder
- Do **NOT** include `src/` directory (source code not needed)
- Do **NOT** include `data/` or `uploads/` directories
- Runtime directories (`data/`, `uploads/`) are created automatically on first run
- Users run `./vibro` from the `vibro-release/` directory
- The binary will create `./data/` and `./uploads/` in that directory
- Users open `http://localhost:8080` in their browser

**Example User Experience:**
```bash
# User extracts vibro-release.zip
cd vibro-release/
./vibro

# Server creates:
# - ./data/contexts.json (empty initially)
# - ./uploads/ (empty initially)

# Server starts on http://localhost:8080
```

---

## 12. API Documentation

### Complete API Reference

```
# Contexts
GET    /api/contexts              → []Context
POST   /api/contexts              → Context (201)
       Body: {name, color, icon}

GET    /api/contexts/:id          → Context
PUT    /api/contexts/:id          → Context
       Body: {name, color, icon}

DELETE /api/contexts/:id          → {success: true} (200)

# Items
POST   /api/contexts/:id/items    → Item (201)
       Body: {type, content}

PUT    /api/items/:id             → Item
       Body: {content, position?}

DELETE /api/items/:id             → {success: true} (200)

# Upload
POST   /api/upload                → {url, filename} (201)
       Body: multipart/form-data (field: "image")
       Max size: 10MB
       Types: PNG, JPG, JPEG, GIF, WEBP
```

---

## 13. Dependencies

### Standard Library Only
```go
import (
    "context"
    "encoding/json"
    "net/http"
    "os"
    "sync"
    "time"
    // ... etc
)
```

### Single External Dependency
```go
// Only for UUID generation
"github.com/google/uuid"
```

**Alternative**: Implement simple UUID v4 generator to remain dependency-free.

---

## 14. Security Considerations

### For Local Single-User App
- **No authentication**: Runs on localhost only
- **No HTTPS**: Local traffic, not needed
- **CORS**: Permissive (localhost development)
- **File uploads**: Type validation, size limits
- **Input validation**: Prevent injection, length limits

### Future Multi-User Considerations
- Add JWT authentication
- Rate limiting per user
- HTTPS/TLS required
- Stricter CORS
- User isolation in storage

---

## 15. Implementation Checklist

### Phase 0: Initial Setup (30 minutes)
- [ ] Initialize Go module in root (`go mod init vibro`)
- [ ] Create `src/` directory for all Go code
- [ ] Create directory structure inside src/ (models/, storage/, handlers/, server/, utils/)
- [ ] Create frontend directory with hello world files
- [ ] Create .gitignore (data/, uploads/)

### Phase 1: Frontend Hello World (30 minutes)
- [ ] Create `frontend/index.html` with hello world
- [ ] Create `frontend/style.css` with basic styling
- [ ] Create `frontend/app.js` with API health check
- [ ] Test that files are created correctly

### Phase 2: Foundation (2-3 hours)
- [ ] Define models (`src/models/models.go`)
- [ ] Implement JSON storage (`src/storage/`)
- [ ] Write storage unit tests

### Phase 3: HTTP Server (2-3 hours)
- [ ] Server setup (`src/server/server.go`)
- [ ] Router configuration (`src/server/router.go`) with static file serving
- [ ] Middleware (logging, recovery, CORS)
- [ ] Response utilities (`src/utils/response.go`)

### Phase 4: Context Handlers (1-2 hours)
- [ ] GET /api/contexts
- [ ] POST /api/contexts
- [ ] GET /api/contexts/:id
- [ ] PUT /api/contexts/:id
- [ ] DELETE /api/contexts/:id
- [ ] Validation

### Phase 5: Item Handlers (1-2 hours)
- [ ] POST /api/contexts/:id/items
- [ ] PUT /api/items/:id
- [ ] DELETE /api/items/:id
- [ ] Position management

### Phase 6: Upload (1 hour)
- [ ] POST /api/upload
- [ ] File validation
- [ ] Unique filename generation
- [ ] Directory creation

### Phase 7: Integration (1 hour)
- [ ] Main entry point (`src/main.go`)
- [ ] Graceful shutdown
- [ ] Runtime directory initialization in `initDirectories()` (creates data/ and uploads/ at runtime)
- [ ] Verify static file serving works (frontend/ must exist in repo)
- [ ] Test build command: `go build -o vibro ./src`
- [ ] Test that data/ and uploads/ are created automatically when binary runs

### Phase 8: Testing (2-3 hours)
- [ ] Unit tests for all components
- [ ] Integration tests for handlers
- [ ] Manual API testing with curl
- [ ] Browser test: visit http://localhost:8080 and verify hello world

### Total Estimate: 11-16 hours

---

## Summary

This backend is designed for:
- ⚡ **Speed**: In-memory storage, no DB overhead
- 🎯 **Simplicity**: Standard library, minimal dependencies
- 🔒 **Reliability**: Atomic writes, error recovery
- 📦 **Portability**: Single binary distribution
- 🧪 **Testability**: Clean interfaces, mockable storage

Perfect for a single-user, local-first application with 3-10 contexts.
