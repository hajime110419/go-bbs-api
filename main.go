// main is a simple API for a bulletin board.
// It provides endpoints to get and create posts.
package main

import (
	"encoding/json"
	"fmt"
	"html" // Used for sanitizing user input.
	"log"
	"net/http"
	"sync" // Used for thread-safe access to the posts slice.

	"github.com/google/uuid" // Used for generating unique IDs for new posts.
)

// Post represents a single post on the bulletin board.
type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var (
	// posts is a slice that stores all the posts in memory.
	posts []Post
	// mu is a RWMutex to ensure safe concurrent access to the posts slice.
	mu sync.RWMutex
)

// sanitize escapes a string to prevent XSS attacks.
func sanitize(s string) string {
	return html.EscapeString(s)
}

// handleRoot is the handler for the root ("/") endpoint.
// It returns a welcome message.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Welcome to the Bulletin Board API! Please use the /posts endpoint.")
}

// handlePosts is the handler for the "/posts" endpoint.
// It dispatches requests to the appropriate handler based on the HTTP method.
func handlePosts(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow cross-origin requests.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight (OPTIONS) requests for CORS.
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

// handleGetPosts handles GET requests to the "/posts" endpoint.
// It retrieves and returns all posts as a JSON array.
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// Use a read lock to allow multiple concurrent reads.
	mu.RLock()
	defer mu.RUnlock()

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("Failed to encode posts to JSON: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// handleCreatePost handles POST requests to the "/posts" endpoint.
// It creates a new post from the request body.
func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var newPost Post

	// Decode the JSON request body into a new Post struct.
	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Assign a new UUID and sanitize the input.
	newPost.ID = uuid.New().String()
	newPost.Title = sanitize((newPost.Title))
	newPost.Content = sanitize(newPost.Content)

	// Use a write lock to ensure exclusive access when modifying the posts slice.
	mu.Lock()
	posts = append(posts, newPost)
	mu.Unlock()

	// Respond with the newly created post.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Printf("Failed to encode new post to JSON: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

func main() {
	// Initialize with a sample post.
	posts = []Post{
		{"00000000-0000-0000-0000-000000000001", "Test Post 1", "This is the first post."},
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/posts", handlePosts)

	port := ":8080"
	fmt.Printf("Starting server on port %s…\n", port)

	// Start the HTTP server.
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
