package utils

import (
	"errors"
	"strings"
	"vibro/src/models"
)

func ValidateCreateContext(req *models.CreateContextRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("context name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("context name too long (max 100 characters)")
	}

	if req.Color != "" && !isValidColor(req.Color) {
		return errors.New("invalid color (must be a valid color name or hex format)")
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

func isValidColor(color string) bool {
	// Valid color names
	validColorNames := map[string]bool{
		"purple":  true,
		"blue":    true,
		"green":   true,
		"orange":  true,
		"red":     true,
		"pink":    true,
		"teal":    true,
		"indigo":  true,
		"cyan":    true,
		"emerald": true,
		"amber":   true,
		"rose":    true,
		"violet":  true,
		"lime":    true,
		"sky":     true,
		"fuchsia": true,
		"slate":   true,
	}

	// Check if it's a valid color name
	if validColorNames[strings.ToLower(color)] {
		return true
	}

	// Check if it's a valid hex color
	return isValidHexColor(color)
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
