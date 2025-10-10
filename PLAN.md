# Context Switching Web Application - Implementation Plan

## Overview
A simple, single-user web application for tracking multiple contexts. Designed for users who switch contexts frequently (every few minutes). Users can create 3-10 contexts, add text, bullet points, and images to each context, and switch between them easily.

## 1. Application Architecture

### Data Model

**Context:**
- ID (string/UUID)
- Name (string)
- Color/Icon (string, for visual distinction)
- CreatedAt (timestamp)
- UpdatedAt (timestamp)

**Item:**
- ID (string/UUID)
- ContextID (string, foreign key)
- Type (string: "text" | "bullet" | "image")
- Content (string)
- Position (integer, for ordering)
- CreatedAt (timestamp)

### Tech Stack
- **Backend:** Go with standard library (net/http)
- **Storage:** JSON files (simple, single-user)
- **Frontend:** Vanilla HTML/CSS/JavaScript
- **No heavy frameworks needed**

---

## 2. Backend Structure (Go)

### Components
- Simple HTTP server with routing
- REST API handlers
- Data persistence layer (JSON file-based)
- Static file server for frontend

### API Endpoints

```
GET    /api/contexts              - List all contexts
POST   /api/contexts              - Create new context
GET    /api/contexts/:id          - Get specific context with items
PUT    /api/contexts/:id          - Update context name
DELETE /api/contexts/:id          - Delete context

POST   /api/contexts/:id/items    - Add item to context
PUT    /api/items/:id             - Update item
DELETE /api/items/:id             - Delete item

POST   /api/upload                - Upload image (save to disk)
```

### Project Repository Structure

```
/
  /src                     - All Go source code
    main.go                - Entry point, server setup
    /server
      server.go            - HTTP server setup
      router.go            - Route registration
      middleware.go        - Logging, CORS, recovery
    /handlers
      context.go           - Context CRUD handlers
      item.go              - Item CRUD handlers
      upload.go            - Image upload handler
    /storage
      storage.go           - Storage interface
      file.go              - JSON file implementation
    /models
      models.go            - Data structures
    /utils
      response.go          - JSON response helpers
      validation.go        - Input validation
  /frontend                - Static files
    index.html             - Single page app
    style.css              - Simple styling
    app.js                 - Frontend logic
  go.mod                   - Go module file (in root)
  go.sum                   - Go dependencies checksum
  .gitignore               - Ignore data/ and uploads/
```

### Runtime Structure (Created when app runs)

```
/
  vibro                    - Compiled Go binary (or vibro.exe on Windows)
  /src                     - Go source code
  /frontend                - Static files (served by Go server)
    index.html
    style.css
    app.js
  /data                    - Created at runtime, git-ignored
    contexts.json          - Data storage
  /uploads                 - Created at runtime, git-ignored
    [image files]          - Uploaded images with unique names
  go.mod                   - Go module file
```

**Notes:**
- The Go server will automatically create `/data` and `/uploads` directories if they don't exist
- Frontend files are served as static assets by the Go HTTP server
- The compiled binary can be distributed as a single executable along with the `/frontend` folder

---

## 3. Frontend Structure

> **Note:** See [LAYOUT.md](LAYOUT.md) for detailed UI mockups and design specifications.

### Key Features

**Context Switching (Multiple Methods):**
- **Command Palette**: `Ctrl+K` opens quick switcher with fuzzy search
- **Arrow Navigation**: `Ctrl+Left/Right` cycles through contexts
- **Number Shortcuts**: `Ctrl+1-9` jumps to specific context
- **Sidebar Menu**: Click `[≡]` button to show all contexts in sidebar
- **Visual Indicator**: Current context shown with color dot + name in header

**Other Features:**
- Auto-save on changes (debounced)
- Drag-and-drop for image upload
- Inline editing for text/bullets
- Color-coded contexts for visual distinction

### JavaScript Structure

**app.js - Main application logic:**
- Context management (switching, creation, deletion)
- Command palette functionality (Ctrl+K with fuzzy search)
- Sidebar toggle and navigation
- Item CRUD operations
- Auto-save functionality (debounced)
- Keyboard shortcuts (Ctrl+K, Ctrl+1-9, Ctrl+Left/Right)
- API communication

---

## 4. Implementation Phases

### Phase 1: Core Backend (Day 1)
- Go server with basic routing
- Context CRUD endpoints
- JSON file storage
- Static file serving

### Phase 2: Basic Frontend (Day 1-2)
- HTML structure
- Context tabs and switcher
- Add/edit/delete contexts
- Display context content

### Phase 3: Item Management (Day 2)
- Add text items
- Add bullet point items
- Edit/delete items
- Backend endpoints for items

### Phase 4: Image Support (Day 3)
- Image upload endpoint
- Frontend drag-and-drop
- Display images in context
- Store images on disk

### Phase 5: Polish (Day 3)
- Keyboard shortcuts
- Auto-save functionality
- Basic CSS styling
- Error handling

---

## 5. Simplest Implementation Decisions

### Storage
**JSON file** (simplest, no database setup needed)
- Single file: `contexts.json`
- Format: Array of contexts with nested items

### Security
**No Authentication:** Single-user, runs locally (localhost:8080)

### State Management
**No Sessions:** Stateless API, frontend manages state

### Image Handling
**Simple file storage:** Save to `/uploads/` with unique filenames

### Frontend Routing
**Single Page:** No routing, all in one HTML file

### Data Persistence
**Auto-save:** Debounced (500ms) to prevent excessive writes

---

## Summary

This plan prioritizes **simplicity** while meeting all core requirements:
- Fast context switching (tabs + keyboard shortcuts)
- Multiple contexts (3-10 supported)
- Text, bullets, and images per context
- Lightweight frontend (no frameworks)
- Simple Go backend (standard library)
- File-based storage (no database complexity)

**Estimated Time:** 2-3 days for a working prototype

---

## Next Steps
1. Set up project structure
2. Implement Phase 1 (Core Backend)
3. Implement Phase 2 (Basic Frontend)
4. Continue through remaining phases
5. Test and iterate
