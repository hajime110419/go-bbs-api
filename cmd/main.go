// Command go-bbs-api is a simple RESTful API for a bulletin board.
// It uses a pure Go SQLite driver to persist post data.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/hajime110419/go-bbs-api/internal/handler"
	"github.com/hajime110419/go-bbs-api/internal/middleware"
	"github.com/hajime110419/go-bbs-api/internal/repository"
	"github.com/hajime110419/go-bbs-api/internal/service"
	"github.com/juju/ratelimit"
)

var (
	db *sql.DB
)

func main() {
	// Initialize the database connection and table schema.
	db = repository.InitDB()
	// Ensure the database connection is closed when the application exits.
	defer db.Close()

	// Initialize the service layer with the database connection.
	postService := service.NewPostService(db)

	// Initialize the handler with the service -> presentation layer
	h := handler.NewPostHandler(postService)

	rate := 2.0
	capacity := int64(2)

	limiterBucket := ratelimit.NewBucketWithRate(rate, capacity)

	limitedHandler := middleware.RateLimiterMiddleware(limiterBucket)(h.HandlePosts)

	http.HandleFunc("/", h.HandleRoot)
	http.HandleFunc("/posts", limitedHandler)

	port := ":8080"
	fmt.Printf("Starting server on port %s…\n", port)

	// Start the HTTP server.
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
