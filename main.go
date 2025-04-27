package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go_crud/database"
	"go_crud/handlers"
)

//go:embed static/index.html
var content embed.FS

func main() {
	// Define command line flags
	persistFlag := flag.Bool("persist", false, "Enable database persistence")
	dbPathFlag := flag.String("dbpath", "database.json", "Path to the database file")
	autoSaveFlag := flag.Bool("autosave", true, "Automatically save changes to file")
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	var (
		db  database.DB
		err error
	)

	// Initialize the database
	if *persistFlag {
		log.Printf("Using persistent database at %s (autosave: %v)", *dbPathFlag, *autoSaveFlag)
		db, err = database.NewPersistentDB(database.Config{
			PersistPath: *dbPathFlag,
			AutoSave:    *autoSaveFlag,
		})
		if err != nil {
			log.Fatalf("Failed to initialize persistent database: %v", err)
		}
	} else {
		log.Println("Using in-memory database (data will be lost when server stops)")
		db = database.NewInMemoryDB()
	}

	// Serve static files at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		data, err := content.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Could not load UI", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	// Register API routes
	http.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetItems(w, r, db)
		case http.MethodPost:
			handlers.CreateItem(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method %s not allowed", r.Method)
		}
	})

	http.HandleFunc("/api/items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetItem(w, r, db)
		case http.MethodPut:
			handlers.UpdateItem(w, r, db)
		case http.MethodDelete:
			handlers.DeleteItem(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method %s not allowed", r.Method)
		}
	})

	// Add endpoint to manually save database
	http.HandleFunc("/api/db/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method %s not allowed", r.Method)
			return
		}

		// Only respond to this endpoint if persistence is enabled
		if *persistFlag {
			if persistentDB, ok := db.(*database.InMemoryDB); ok {
				if err := persistentDB.SaveToFile(*dbPathFlag); err != nil {
					http.Error(w, fmt.Sprintf("Failed to save database: %v", err), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"status":"success","message":"Database saved to %s"}`, *dbPathFlag)
			} else {
				http.Error(w, "Database does not support persistence", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Persistence not enabled", http.StatusBadRequest)
		}
	})

	// Start the server
	serverAddr := ":" + *port
	serverCh := make(chan struct{})
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		log.Printf("Visit http://localhost%s to access the CRUD UI", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}
		close(serverCh)
	}()

	// Handle graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		log.Println("Shutting down server...")

		// If persistence is enabled, save the database before exiting
		if *persistFlag {
			if persistentDB, ok := db.(*database.InMemoryDB); ok {
				log.Println("Saving database before exit...")
				if err := persistentDB.SaveToFile(*dbPathFlag); err != nil {
					log.Printf("Failed to save database: %v", err)
				} else {
					log.Printf("Database saved to %s", *dbPathFlag)
				}
			}
		}
	case <-serverCh:
		log.Println("Server stopped")
	}
}
