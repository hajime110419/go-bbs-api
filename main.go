// Command go-bbs-api is a simple RESTful API for a bulletin board.
// It uses a pure Go SQLite driver to persist post data.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html" // Used for sanitizing user input to prevent XSS.
	"log"
	"net/http"

	"github.com/google/uuid" // Used for generating unique IDs for new posts.
	"github.com/juju/ratelimit"
	_ "modernc.org/sqlite" // Pure Go SQLite driver, CGO-free.
)

// Post represents a single entry on the bulletin board.
type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var (
	// db holds the application's database connection pool. It is initialized
	// in main() and shared across all handlers.
	db *sql.DB
)

// createTable ensures the necessary database schema (the 'posts' table) exists.
// It accepts a database connection pool and can be used for both the main application
// and for setting up test databases.
func createTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
		"id" TEXT NOT NULL PRIMARY KEY,
		"title" TEXT,
		"content" TEXT
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		// Since this is a critical startup step, we terminate if it fails.
		log.Fatalf("Failed to create table: %v", err)
	}
}

// initDB now focuses solely on opening the main application's database file
// and then calls createTable to set up the schema.
func initDB() {
	var err error
	// The driver name "sqlite" is registered by the blank import of modernc.org/sqlite.
	db, err = sql.Open("sqlite", "./bulletinboard.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Call the separated function to create the table schema.
	createTable(db)
}

// sanitize escapes potentially harmful characters from a string to prevent
// Cross-Site Scripting (XSS) attacks.
func sanitize(s string) string {
	return html.EscapeString(s)
}

// rateLimiterMiddleware returns an HTTP middleware that applies a rate limit
// using the juju/ratelimit token bucket.
func rateLimiterMiddleware(bucket *ratelimit.Bucket) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if a token is available. TakeAvailable(1) attempts to consume 1 token
			// immediately and returns 0 if none are avaliable.
			if bucket.TakeAvailable(1) == 0 {
				// If rate limit is exceeded, return 429 Too Many Requests.
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				http.Error(w, `{"error": "Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}
			// If a token is consumed, proceed to the next handler.
			next(w, r)
		}
	}
}

// handleRoot is the handler for the root ("/") endpoint.
// It returns a simple welcome message.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Welcome to the Bulletin Board API! Please use the /posts endpoint.")
}

// handlePosts routes requests for the "/posts" endpoint based on the HTTP method.
// It also handles CORS preflight (OPTIONS) requests.
func handlePosts(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow cross-origin requests from web browsers.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle CORS preflight requests.
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch r.Method {
	case "GET":
		handleGetPosts(w, r)
	case "POST":
		handleCreatePost(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// handleGetPosts handles GET requests to "/posts". It retrieves all posts from
// the database, ordered by creation time (descending), and returns them as a JSON array.
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// "rowid" is an implicit auto-incrementing column in SQLite. Ordering by it
	// in descending order retrieves the most recent posts first.
	rows, err := db.Query("SELECT id, title, content FROM posts ORDER BY rowid DESC")
	if err != nil {
		log.Printf("Failed to query posts from database: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content); err != nil {
			log.Printf("Failed to scan row: %v", err)
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("Failed to encode posts to JSON: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// handleCreatePost handles POST requests to "/posts". It decodes a new post
// from the request body, assigns a unique ID, sanitizes the input, and inserts
// it into the database. It returns the newly created post as JSON.
func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var newPost Post

	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Assign a new universally unique identifier (UUID).
	newPost.ID = uuid.New().String()
	// Sanitize user-provided title and content to prevent XSS.
	newPost.Title = sanitize(newPost.Title)
	newPost.Content = sanitize(newPost.Content)

	// Use a prepared statement to prevent SQL injection vulnerabilities.
	stmt, err := db.Prepare("INSERT INTO posts(id, title, content) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare statement: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPost.ID, newPost.Title, newPost.Content)
	if err != nil {
		log.Printf("Failed to insert post into database: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Printf("Failed to encode new post to JSON: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

func main() {
	// Initialize the database connection and table schema.
	initDB()
	// Ensure the database connection is closed when the application exits.
	defer db.Close()

	rate := 2.0
	capacity := int64(2)

	limiterBucket := ratelimit.NewBucketWithRate(rate, capacity)

	limitedHandler := rateLimiterMiddleware(limiterBucket)(handlePosts)

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/posts", limitedHandler)

	port := ":8080"
	fmt.Printf("Starting server on port %s…\n", port)

	// Start the HTTP server.
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
