package server

import (
	"net/http"
	"vibro/src/handlers"
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
