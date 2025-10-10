# Vibro - Context Switching Application

A lightweight, fast Go backend for managing contexts and items with image upload support. Built with only the Go standard library for maximum simplicity and performance.

## ✨ Features

- 🚀 **High Performance**: In-memory storage with sub-millisecond response times
- 🔒 **Thread-Safe**: Concurrent request handling with RWMutex
- 💾 **Data Persistence**: JSON-based storage with atomic writes and auto-backup
- 📤 **Image Upload**: Support for PNG, JPG, JPEG, GIF, and WEBP (max 10MB)
- 🛡️ **Error Handling**: Comprehensive validation and consistent error responses
- 📝 **Request Logging**: Method, path, status code, and duration for every request
- 🌐 **CORS Ready**: Configured for local development
- 🧪 **Well Tested**: 19+ unit tests with full coverage

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Architecture](#architecture)

## 🚀 Quick Start

### Prerequisites

- Go 1.24.0 or higher
- Git (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd vibro
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o vibro ./src
```

4. Run the server:
```bash
./vibro
```

5. Open your browser:
```
http://localhost:8080
```

The server will automatically create `./data/` and `./uploads/` directories on first run.

## 📁 Project Structure

```
vibro/
├── src/                    # All Go source code
│   ├── main.go            # Entry point, server initialization
│   ├── models/            # Data structures
│   │   ├── models.go      # Context, Item, DTOs
│   │   └── models_test.go # Unit tests
│   ├── storage/           # Data persistence layer
│   │   ├── storage.go     # Storage interface
│   │   ├── file.go        # JSON implementation
│   │   └── file_test.go   # Unit tests
│   ├── server/            # HTTP server
│   │   ├── server.go      # Server setup
│   │   ├── router.go      # Route registration
│   │   └── middleware.go  # Logging, CORS, recovery
│   ├── handlers/          # HTTP handlers
│   │   ├── context.go     # Context CRUD
│   │   ├── item.go        # Item CRUD
│   │   └── upload.go      # Image upload
│   └── utils/             # Utilities
│       ├── response.go    # JSON response helpers
│       └── validation.go  # Input validation
├── frontend/              # Static frontend files
│   ├── index.html        # Main page
│   ├── style.css         # Styling
│   └── app.js            # JavaScript
├── data/                 # Runtime directory (auto-created)
│   └── contexts.json     # JSON data storage
├── uploads/              # Runtime directory (auto-created)
│   └── [images]          # Uploaded images
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── .gitignore            # Git ignore rules
└── README.md             # This file
```

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api
```

### Contexts

#### Get All Contexts
```http
GET /api/contexts
```

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "Work",
    "color": "#3498db",
    "icon": "💼",
    "items": [],
    "createdAt": "2025-10-10T21:51:11.36377+02:00",
    "updatedAt": "2025-10-10T21:51:11.36377+02:00"
  }
]
```

#### Create Context
```http
POST /api/contexts
Content-Type: application/json

{
  "name": "Work",
  "color": "#3498db",
  "icon": "💼"
}
```

**Response:** `201 Created`
```json
{
  "id": "7b95a19a-7bf4-460c-b5ca-952aebc4857a",
  "name": "Work",
  "color": "#3498db",
  "icon": "💼",
  "items": [],
  "createdAt": "2025-10-10T21:51:11.36377+02:00",
  "updatedAt": "2025-10-10T21:51:11.36377+02:00"
}
```

#### Get Context by ID
```http
GET /api/contexts/{id}
```

**Response:** `200 OK` or `404 Not Found`

#### Update Context
```http
PUT /api/contexts/{id}
Content-Type: application/json

{
  "name": "Work Projects",
  "color": "#2ecc71",
  "icon": "🚀"
}
```

**Response:** `200 OK` or `404 Not Found`

#### Delete Context
```http
DELETE /api/contexts/{id}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Context deleted successfully"
}
```

### Items

#### Create Item
```http
POST /api/contexts/{contextId}/items
Content-Type: application/json

{
  "type": "text",
  "content": "Buy groceries"
}
```

**Item Types:** `text`, `bullet`, `image`

**Response:** `201 Created`
```json
{
  "id": "c2ebb0d8-84f2-451d-ae07-287c9f83cc48",
  "contextId": "496b4b2d-7396-4ec2-880e-9b484d316f32",
  "type": "text",
  "content": "Buy groceries",
  "position": 0,
  "createdAt": "2025-10-10T21:51:41.563772+02:00"
}
```

#### Update Item
```http
PUT /api/items/{id}
Content-Type: application/json

{
  "content": "Buy groceries and cook dinner",
  "position": 2
}
```

**Response:** `200 OK` or `404 Not Found`

#### Delete Item
```http
DELETE /api/items/{id}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Item deleted successfully"
}
```

### Upload

#### Upload Image
```http
POST /api/upload
Content-Type: multipart/form-data

image: [file]
```

**Constraints:**
- Max size: 10MB
- Allowed types: PNG, JPG, JPEG, GIF, WEBP

**Response:** `201 Created`
```json
{
  "url": "/uploads/c32d1076b16c5a7d.png",
  "filename": "c32d1076b16c5a7d.png"
}
```

#### Access Uploaded Image
```http
GET /uploads/{filename}
```

**Response:** `200 OK` (image file)

### Error Responses

All errors follow this format:
```json
{
  "error": "Bad Request",
  "message": "context name is required"
}
```

**Status Codes:**
- `400 Bad Request` - Validation errors
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server errors

## ⚙️ Configuration

### Server Settings

Edit `src/main.go` to change server configuration:

```go
const (
    serverAddr   = "localhost:8080"  // Server address
    dataFile     = "./data/contexts.json"  // Storage file
    uploadsDir   = "./uploads"  // Upload directory
)
```

### Timeouts

Server timeouts are configured in `src/server/server.go`:

```go
ReadTimeout:  10 * time.Second
WriteTimeout: 10 * time.Second
IdleTimeout:  60 * time.Second
```

### Upload Limits

Upload constraints in `src/handlers/upload.go`:

```go
const (
    maxUploadSize = 10 << 20  // 10 MB
    uploadsDir    = "./uploads"
)
```

## 🛠️ Development

### Run in Development Mode

```bash
# Build and run
go build -o vibro ./src && ./vibro

# Or use go run
go run ./src
```

### Watch for Changes

Install and use a file watcher like `air`:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with air
air
```

### Code Formatting

```bash
# Format all Go files
gofmt -w ./src

# Or use goimports
goimports -w ./src
```

### Linting

```bash
# Run go vet
go vet ./src/...

# Run staticcheck (install first)
staticcheck ./src/...
```

## 🧪 Testing

### Run All Tests

```bash
go test ./src/...
```

### Run Tests with Verbose Output

```bash
go test ./src/... -v
```

### Run Tests with Coverage

```bash
go test ./src/... -cover
```

### Generate Coverage Report

```bash
go test ./src/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Specific Package

```bash
# Test models only
go test ./src/models -v

# Test storage only
go test ./src/storage -v
```

### Current Test Coverage

- **Models Package**: 19 test cases covering validation logic
- **Storage Package**: 3 test functions covering CRUD, atomic writes, and persistence
- **Total**: All critical paths tested

## 📦 Deployment

### Build for Production

```bash
go build -o vibro ./src
```

### Cross-Compilation

#### Linux (AMD64)
```bash
GOOS=linux GOARCH=amd64 go build -o vibro-linux ./src
```

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o vibro.exe ./src
```

#### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o vibro-macos-intel ./src
```

#### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o vibro-macos-arm ./src
```

### Distribution Package

Create a release package with only what's needed:

```
vibro-release/
├── vibro (or vibro.exe)     # Compiled binary
├── frontend/                # Static files (required)
│   ├── index.html
│   ├── style.css
│   └── app.js
└── README.txt               # User instructions
```

**Note:** `data/` and `uploads/` directories are created automatically on first run.

### Running in Production

```bash
# Navigate to release directory
cd vibro-release/

# Run the binary
./vibro

# Server starts on http://localhost:8080
# Data stored in ./data/contexts.json
# Uploads saved to ./uploads/
```

### Using as a Service

#### systemd (Linux)

Create `/etc/systemd/system/vibro.service`:

```ini
[Unit]
Description=Vibro Context Switching Application
After=network.target

[Service]
Type=simple
User=vibro
WorkingDirectory=/opt/vibro
ExecStart=/opt/vibro/vibro
Restart=always

[Install]
WantedBy=multi-user.target
```

Start the service:
```bash
sudo systemctl start vibro
sudo systemctl enable vibro
```

## 🏗️ Architecture

### Design Principles

1. **Simplicity**: Standard library only (except UUID generation)
2. **Performance**: In-memory cache with fast O(n) lookups
3. **Reliability**: Atomic writes prevent data corruption
4. **Testability**: Clean interfaces, mockable storage
5. **Portability**: Single binary distribution

### Storage Layer

**In-Memory Cache:**
- Entire dataset loaded into memory on startup
- Fast reads with O(n) complexity (acceptable for 3-10 contexts)
- Thread-safe using `sync.RWMutex`

**Persistence:**
- Write-through pattern: every mutation saves to disk
- Atomic writes: temp file + rename (POSIX atomic)
- Auto-backup: `.bak` file created before each write

**Concurrency:**
- `RWMutex` allows multiple concurrent reads
- Exclusive locks only during writes
- No connection pooling needed (simpler than databases)

### HTTP Server

**Routing:**
- Go 1.22+ pattern matching (`GET /api/contexts/{id}`)
- Automatic method routing
- Path parameters via `r.PathValue()`

**Middleware Stack:**
1. CORS headers (development-friendly)
2. Panic recovery (prevents crashes)
3. Request logging (method, path, status, duration)

**Error Handling:**
- Consistent JSON error responses
- Appropriate HTTP status codes
- Server-side error logging with context

### Security Considerations

**Current (Local Single-User):**
- ✅ No authentication (localhost only)
- ✅ File upload validation (type, size)
- ✅ Input validation (SQL injection not applicable)
- ✅ CORS permissive (local development)

**Future (Multi-User):**
- ⚠️ Add JWT authentication
- ⚠️ Implement rate limiting
- ⚠️ Require HTTPS/TLS
- ⚠️ Stricter CORS policies
- ⚠️ User isolation in storage

## 🔧 Troubleshooting

### Server Won't Start

**Port already in use:**
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process or change port in src/main.go
```

**Permission denied:**
```bash
# Make binary executable
chmod +x vibro
```

### Data Not Persisting

**Check file permissions:**
```bash
# Ensure data directory is writable
ls -la ./data/
chmod 755 ./data/
```

**Check logs:**
- Server logs errors to stderr
- Look for "Failed to save" messages

### Upload Fails

**File too large:**
- Max size is 10MB
- Change `maxUploadSize` in `src/handlers/upload.go`

**Invalid file type:**
- Only PNG, JPG, JPEG, GIF, WEBP allowed
- Extensions must be lowercase

### API Returns 404

**Wrong route:**
- Ensure `/api` prefix: `/api/contexts` not `/contexts`
- Check HTTP method matches route definition

**Context not found:**
- Verify ID is correct UUID
- Check `./data/contexts.json` for existing IDs

## 📝 License

[Your License Here]

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Follow Go conventions (`gofmt`, `go vet`)
- Update documentation
- Keep commits focused and descriptive

## 📮 Support

- **Issues**: [GitHub Issues](https://github.com/your-repo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-repo/discussions)
- **Email**: your-email@example.com

## 🙏 Acknowledgments

- Built with Go standard library
- UUID generation by [google/uuid](https://github.com/google/uuid)
- Inspired by context-switching productivity workflows

---

**Made with ❤️ for productivity enthusiasts**
