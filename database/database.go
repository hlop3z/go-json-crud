package database

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"go_crud/models"
)

var (
	ErrNotFound  = errors.New("item not found")
	ErrInvalidID = errors.New("invalid item ID")
)

// DB is an interface for database operations
type DB interface {
	GetAll() ([]*models.Item, error)
	Get(id string) (*models.Item, error)
	Create(item *models.Item) error
	Update(item *models.Item) error
	Delete(id string) error
	SaveToFile(filepath string) error
	LoadFromFile(filepath string) error
}

// InMemoryDB is a simple in-memory database implementation
type InMemoryDB struct {
	items       map[string]*models.Item
	mutex       sync.RWMutex
	persistPath string
	autoSave    bool
}

// Config holds configuration options for the database
type Config struct {
	PersistPath string
	AutoSave    bool
}

// NewInMemoryDB creates a new in-memory database
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		items: make(map[string]*models.Item),
	}
}

// NewPersistentDB creates a new database with persistence enabled
func NewPersistentDB(config Config) (*InMemoryDB, error) {
	db := &InMemoryDB{
		items:       make(map[string]*models.Item),
		persistPath: config.PersistPath,
		autoSave:    config.AutoSave,
	}

	// If the file exists, load items from it
	if _, err := os.Stat(config.PersistPath); err == nil {
		if err := db.LoadFromFile(config.PersistPath); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// SaveToFile saves the database contents to a JSON file
func (db *InMemoryDB) SaveToFile(filepath string) error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// Convert map to slice for easier JSON serialization
	items := make([]*models.Item, 0, len(db.items))
	for _, item := range db.items {
		items = append(items, item)
	}

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode items as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

// LoadFromFile loads the database contents from a JSON file
func (db *InMemoryDB) LoadFromFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var items []*models.Item
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		// Handle empty file case
		if err == io.EOF {
			return nil
		}
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Clear current items and load from file
	db.items = make(map[string]*models.Item)
	for _, item := range items {
		db.items[item.ID] = item
	}

	return nil
}

// GetAll returns all items in the database
func (db *InMemoryDB) GetAll() ([]*models.Item, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	items := make([]*models.Item, 0, len(db.items))
	for _, item := range db.items {
		items = append(items, item)
	}
	return items, nil
}

// Get returns an item by ID
func (db *InMemoryDB) Get(id string) (*models.Item, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	item, exists := db.items[id]
	if !exists {
		return nil, ErrNotFound
	}
	return item, nil
}

// Create adds a new item to the database
func (db *InMemoryDB) Create(item *models.Item) error {
	if item.ID == "" {
		return ErrInvalidID
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.items[item.ID]; exists {
		return errors.New("item with this ID already exists")
	}

	db.items[item.ID] = item

	// Auto-save if enabled
	if db.autoSave && db.persistPath != "" {
		go db.SaveToFile(db.persistPath) // Save in background
	}

	return nil
}

// Update modifies an existing item
func (db *InMemoryDB) Update(item *models.Item) error {
	if item.ID == "" {
		return ErrInvalidID
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.items[item.ID]; !exists {
		return ErrNotFound
	}

	db.items[item.ID] = item

	// Auto-save if enabled
	if db.autoSave && db.persistPath != "" {
		go db.SaveToFile(db.persistPath) // Save in background
	}

	return nil
}

// Delete removes an item from the database
func (db *InMemoryDB) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.items[id]; !exists {
		return ErrNotFound
	}

	delete(db.items, id)

	// Auto-save if enabled
	if db.autoSave && db.persistPath != "" {
		go db.SaveToFile(db.persistPath) // Save in background
	}

	return nil
}
