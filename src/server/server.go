package server

import (
	"context"
	"net/http"
	"time"
	"vibro/src/storage"
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
