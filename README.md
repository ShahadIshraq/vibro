# Vibro

A simple Go backend for managing contexts and items with image upload support.

## Quick Start

Requirements: Go 1.24+

```bash
go build -o vibro ./src
./vibro

# Or run on a custom port
./vibro --port 3000
```

Open http://localhost:8080 (or your custom port)

## API

Base URL: `http://localhost:8080/api`

### Contexts

```http
GET    /api/contexts          # List all
POST   /api/contexts          # Create
GET    /api/contexts/{id}     # Get one
PUT    /api/contexts/{id}     # Update
DELETE /api/contexts/{id}     # Delete
```

### Items

```http
POST   /api/contexts/{contextId}/items  # Create item
PUT    /api/items/{id}                  # Update item
DELETE /api/items/{id}                  # Delete item
```

### Upload

```http
POST   /api/upload            # Upload image (max 10MB)
GET    /uploads/{filename}    # Serve uploaded image
```

Supported formats: PNG, JPG, JPEG, GIF, WEBP

## Project Structure

```
src/
├── main.go            # Entry point
├── models/            # Data structures
├── storage/           # JSON file persistence
├── server/            # HTTP server, routing, middleware
├── handlers/          # API handlers
└── utils/             # Helpers
frontend/              # Static HTML/CSS/JS
```

## Testing

```bash
go test ./src/...
```

## Configuration

### Command-line flags

- `--port` - Port to run the server on (default: 8080)

Example:
```bash
./vibro --port 3000
```

### File paths

Edit constants in `src/main.go`:
- Data file location (default: ./data/contexts.json)
- Upload directory (default: ./uploads)

## Deployment

### Build

```bash
# Current platform
go build -o vibro ./src

# Linux
GOOS=linux GOARCH=amd64 go build -o vibro-linux ./src

# Windows
GOOS=windows GOARCH=amd64 go build -o vibro.exe ./src
```

### Deploy

The binary expects this directory structure at runtime:

```
vibro/
├── vibro              # The compiled binary
└── frontend/          # Required - static files served by the app
    ├── index.html
    ├── style.css
    └── app.js
```

Run the binary from the directory containing `frontend/`:

```bash
./vibro

# Or specify a custom port
./vibro --port 8000
```

The app will create `data/` and `uploads/` directories automatically on first run.
