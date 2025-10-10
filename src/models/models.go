package models

import (
	"errors"
	"strings"
	"time"
)

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

// Validate validates a Context
func (c *Context) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return errors.New("context ID is required")
	}

	if strings.TrimSpace(c.Name) == "" {
		return errors.New("context name is required")
	}

	if len(c.Name) > 100 {
		return errors.New("context name too long (max 100 characters)")
	}

	if c.Color != "" && !isValidHexColor(c.Color) {
		return errors.New("invalid color format (use hex: #RRGGBB)")
	}

	return nil
}

// Validate validates an Item
func (i *Item) Validate() error {
	if strings.TrimSpace(i.ID) == "" {
		return errors.New("item ID is required")
	}

	if strings.TrimSpace(i.ContextID) == "" {
		return errors.New("item context ID is required")
	}

	if i.Type != ItemTypeText && i.Type != ItemTypeBullet && i.Type != ItemTypeImage {
		return errors.New("invalid item type")
	}

	if strings.TrimSpace(i.Content) == "" {
		return errors.New("item content is required")
	}

	if i.Position < 0 {
		return errors.New("item position must be non-negative")
	}

	return nil
}

// isValidHexColor checks if a color string is a valid hex color
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
