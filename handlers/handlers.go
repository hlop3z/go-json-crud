package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go_crud/database"
	"go_crud/models"
)

// Helper to get ID from URL
func getIDFromURL(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 3 {
		return parts[3] // Format is /api/items/{id}
	}
	return ""
}

// Helper to send JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error encoding response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper to send error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// GetItems retrieves all items
func GetItems(w http.ResponseWriter, r *http.Request, db database.DB) {
	items, err := db.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving items")
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

// GetItem retrieves a specific item by ID
func GetItem(w http.ResponseWriter, r *http.Request, db database.DB) {
	id := getIDFromURL(r)
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item, err := db.Get(id)
	if err != nil {
		if err == database.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Item not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving item")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

// CreateItem adds a new item
func CreateItem(w http.ResponseWriter, r *http.Request, db database.DB) {
	var item models.Item
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set timestamps
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := db.Create(&item); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating item")
		return
	}

	respondWithJSON(w, http.StatusCreated, item)
}

// UpdateItem modifies an existing item
func UpdateItem(w http.ResponseWriter, r *http.Request, db database.DB) {
	id := getIDFromURL(r)
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	existingItem, err := db.Get(id)
	if err != nil {
		if err == database.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Item not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving item")
		}
		return
	}

	var updatedItem models.Item
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &updatedItem); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Preserve ID and creation timestamp
	updatedItem.ID = id
	updatedItem.CreatedAt = existingItem.CreatedAt
	updatedItem.UpdatedAt = time.Now()

	if err := db.Update(&updatedItem); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating item")
		return
	}

	respondWithJSON(w, http.StatusOK, updatedItem)
}

// DeleteItem removes an item
func DeleteItem(w http.ResponseWriter, r *http.Request, db database.DB) {
	id := getIDFromURL(r)
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	if err := db.Delete(id); err != nil {
		if err == database.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Item not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error deleting item")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
