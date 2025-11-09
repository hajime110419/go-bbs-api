package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hajime110419/go-bbs-api/internal/service"
)

type PostHandler struct {
	service *service.PostService
}

// NewPostHandler creates a new PostHandler with the given PostService.
func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{service: svc}
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

// HandleGetPosts handles GET requests to "/posts". It retrieves all posts
// through the service layer and returns them as a JSON array.
func (h *PostHandler) HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetAllPosts()
	if err != nil {
		log.Printf("Failed to query posts from database: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("Failed to encode posts to JSON: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// HandleCreatePost handles POST requests to "/posts". It decodes a new post
// from the request body, delegates creation to the service layer,
// and returns the newly created post as JSON.
func (h *PostHandler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	post, err := h.service.CreatePost(input.Title, input.Content)
	if err != nil {
		log.Printf("Failed to insert post into database: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
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
