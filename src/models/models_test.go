package models

import (
	"testing"
	"time"
)

func TestContextValidate(t *testing.T) {
	tests := []struct {
		name    string
		context Context
		wantErr bool
	}{
		{
			name: "valid context",
			context: Context{
				ID:        "123",
				Name:      "Work",
				Color:     "#FF5733",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			context: Context{
				Name:      "Work",
				Color:     "#FF5733",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing name",
			context: Context{
				ID:        "123",
				Color:     "#FF5733",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "valid color name",
			context: Context{
				ID:        "123",
				Name:      "Work",
				Color:     "red",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid color",
			context: Context{
				ID:        "123",
				Name:      "Work",
				Color:     "notacolor",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "name too long",
			context: Context{
				ID:        "123",
				Name:      string(make([]byte, 101)),
				Color:     "#FF5733",
				Icon:      "💼",
				Items:     []Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.context.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Context.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestItemValidate(t *testing.T) {
	tests := []struct {
		name    string
		item    Item
		wantErr bool
	}{
		{
			name: "valid text item",
			item: Item{
				ID:        "456",
				ContextID: "123",
				Type:      ItemTypeText,
				Content:   "Some text",
				Position:  0,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid bullet item",
			item: Item{
				ID:        "456",
				ContextID: "123",
				Type:      ItemTypeBullet,
				Content:   "Bullet point",
				Position:  1,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			item: Item{
				ContextID: "123",
				Type:      ItemTypeText,
				Content:   "Some text",
				Position:  0,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing context ID",
			item: Item{
				ID:        "456",
				Type:      ItemTypeText,
				Content:   "Some text",
				Position:  0,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			item: Item{
				ID:        "456",
				ContextID: "123",
				Type:      "invalid",
				Content:   "Some text",
				Position:  0,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty content",
			item: Item{
				ID:        "456",
				ContextID: "123",
				Type:      ItemTypeText,
				Content:   "",
				Position:  0,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative position",
			item: Item{
				ID:        "456",
				ContextID: "123",
				Type:      ItemTypeText,
				Content:   "Some text",
				Position:  -1,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Item.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{"valid uppercase", "#FF5733", true},
		{"valid lowercase", "#ff5733", true},
		{"valid mixed case", "#Ff5733", true},
		{"missing hash", "FF5733", false},
		{"too short", "#FFF", false},
		{"too long", "#FF57331", false},
		{"invalid characters", "#GG5733", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidHexColor(tt.color); got != tt.want {
				t.Errorf("isValidHexColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
