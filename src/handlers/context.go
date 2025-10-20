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
			ID:          uuid.New().String(),
			Name:        req.Name,
			Color:       req.Color,
			Icon:        req.Icon,
			Description: req.Description,
			Items:       make([]models.Item, 0),
			Notes:       req.Notes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Initialize empty notes array if nil
		if ctx.Notes == nil {
			ctx.Notes = make([]models.Note, 0)
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

		// Validate input
		createReq := &models.CreateContextRequest{
			Name:        req.Name,
			Color:       req.Color,
			Icon:        req.Icon,
			Description: req.Description,
			Notes:       req.Notes,
		}
		if err := utils.ValidateCreateContext(createReq); err != nil {
			utils.RespondError(w, http.StatusBadRequest, err.Error(), nil)
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
		ctx.Description = req.Description
		ctx.Notes = req.Notes
		ctx.UpdatedAt = time.Now()

		// Initialize empty notes array if nil
		if ctx.Notes == nil {
			ctx.Notes = make([]models.Note, 0)
		}

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

// ReorderContexts updates the order of contexts
func ReorderContexts(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Order []string `json:"order"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if len(req.Order) == 0 {
			utils.RespondError(w, http.StatusBadRequest, "Order array cannot be empty", nil)
			return
		}

		// Get all contexts
		contexts, err := store.GetAllContexts()
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch contexts", err)
			return
		}

		// Create a map for quick context lookup
		contextMap := make(map[string]*models.Context)
		for i := range contexts {
			contextMap[contexts[i].ID] = &contexts[i]
		}

		// Reorder contexts based on the provided order
		reorderedContexts := make([]models.Context, 0, len(req.Order))
		for _, id := range req.Order {
			if ctx, exists := contextMap[id]; exists {
				reorderedContexts = append(reorderedContexts, *ctx)
			}
		}

		// Update the order in storage
		if err := store.ReorderContexts(reorderedContexts); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to reorder contexts", err)
			return
		}

		utils.RespondJSON(w, http.StatusOK, models.SuccessResponse{
			Success: true,
			Message: "Contexts reordered successfully",
		})
	}
}
