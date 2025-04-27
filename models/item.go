package models

import (
	"time"
)

// Item represents a basic item in our CRUD application
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewItem creates a new item with the given name and description
func NewItem(id, name, description string) *Item {
	now := time.Now()
	return &Item{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
