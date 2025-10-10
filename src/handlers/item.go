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
