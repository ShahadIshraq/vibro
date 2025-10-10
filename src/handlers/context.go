package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"vibro/src/models"
	"vibro/src/storage"
	"vibro/src/utils"

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
