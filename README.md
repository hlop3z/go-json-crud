# Go CRUD Application

A simple CRUD (Create, Read, Update, Delete) application built with pure Go and no external dependencies. It provides a RESTful API for managing items.

## Features

- In-memory database with thread-safe operations
- Optional JSON-based persistence
- JSON API for item management
- Web interface for testing CRUD operations
- Pure Go implementation with no external dependencies
- Proper error handling and response codes

## Running the Application

To run the application:

```bash
cd go_crud
go run main.go
```

To run the application (enable database persistence):

```bash
go run main.go --persist
```

The server will start on port 8080 with an in-memory database (data will be lost when the server stops).

### Command Line Options

You can use the following command line options:

- `-persist`: Enable database persistence (default: false)
- `-dbpath`: Path to the database file (default: "items.json")
- `-autosave`: Automatically save changes to file (default: true)
- `-port`: Port to run the server on (default: "8080")

Example with persistence enabled:

```bash
go run main.go -persist -dbpath data.json
```

## Database Persistence

When persistence is enabled, the application will:

1. Load existing items from the database file at startup
2. Save items to the file when changes occur (if autosave is enabled)
3. Save items when the server shuts down
4. Provide an API endpoint to manually save the database

You can manually save the database at any time by clicking the "Save Database" button in the web interface or by making a POST request to `/api/db/save`.

## Web Interface

Visit <http://localhost:8080> in your browser to access the web interface. The interface allows you to:

- View all items
- Create new items
- Update existing items
- Delete items
- Save the database (when persistence is enabled)

The web interface is a simple HTML/JS application that makes API calls to the backend.

## API Endpoints

### Get All Items

```text
GET /api/items
```

### Get Single Item

```text
GET /api/items/{id}
```

### Create Item

```text
POST /api/items
Content-Type: application/json

{
  "id": "1",
  "name": "Example Item",
  "description": "This is an example item"
}
```

### Update Item

```text
PUT /api/items/{id}
Content-Type: application/json

{
  "name": "Updated Item",
  "description": "This item has been updated"
}
```

### Delete Item

```text
DELETE /api/items/{id}
```

### Save Database (when persistence is enabled)

```text
POST /api/db/save
```

## Example Usage with cURL

### Create an item

```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"id":"1","name":"Test Item","description":"A test item"}'
```

### Get all items

```bash
curl http://localhost:8080/api/items
```

### Get a specific item

```bash
curl http://localhost:8080/api/items/1
```

### Update an item

```bash
curl -X PUT http://localhost:8080/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Item","description":"This item has been updated"}'
```

### Delete an item

```bash
curl -X DELETE http://localhost:8080/api/items/1
```

### Save database

```bash
curl -X POST http://localhost:8080/api/db/save
```
