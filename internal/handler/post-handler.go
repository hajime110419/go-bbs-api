package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid" // Used for generating unique IDs for new posts.
	"github.com/hajime110419/go-bbs-api/internal/models"
	"github.com/hajime110419/go-bbs-api/internal/utils"
)

type PostHandler struct {
	DB *sql.DB
}

// HandlePosts routes requests for the "/posts" endpoint based on the HTTP method.
// It also handles CORS preflight (OPTIONS) requests.
func (h *PostHandler) HandlePosts(w http.ResponseWriter, r *http.Request) {
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
		h.HandleGetPosts(w, r)
	case "POST":
		h.HandleCreatePost(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleGetPosts handles GET requests to "/posts". It retrieves all posts from
// the database, ordered by creation time (descending), and returns them as a JSON array.
func (h *PostHandler) HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	// "rowid" is an implicit auto-incrementing column in SQLite. Ordering by it
	// in descending order retrieves the most recent posts first.
	rows, err := h.DB.Query("SELECT id, title, content FROM posts ORDER BY rowid DESC")
	if err != nil {
		log.Printf("Failed to query posts from database: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		var p models.Post
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

// HandleCreatePost handles POST requests to "/posts". It decodes a new post
// from the request body, assigns a unique ID, sanitizes the input, and inserts
// it into the database. It returns the newly created post as JSON.
func (h *PostHandler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	var newPost models.Post

	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Assign a new universally unique identifier (UUID).
	newPost.ID = uuid.New().String()
	// Sanitize user-provided title and content to prevent XSS.
	newPost.Title = utils.Sanitize(newPost.Title)
	newPost.Content = utils.Sanitize(newPost.Content)

	// Use a prepared statement to prevent SQL injection vulnerabilities.
	stmt, err := h.DB.Prepare("INSERT INTO posts(id, title, content) VALUES(?, ?, ?)")
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

// HandleRoot is the handler for the root ("/") endpoint.
// It returns a simple welcome message.
func (h *PostHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Welcome to the Bulletin Board API! Please use the /posts endpoint.")
}
