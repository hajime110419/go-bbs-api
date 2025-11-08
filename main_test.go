package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // Ensure the driver is imported for tests
)

func setupTestDB(t *testing.T) func() {
	var err error
	// Use ":memory:" to create a private, in-memory database for each test.
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Call the new, reusable createTable function directly.
	// This correctly applies the schema to our in-memory database
	// without touching the file-based one.
	createTable(db)

	// Return a teardown function to be called by the test.
	teardown := func() {
		db.Close()
	}
	return teardown
}

// TestGetEndpoints tests GET requests for the root and /posts endpoints.
func TestGetEndpoints(t *testing.T) {
	// Setup the test database and schedule its cleanup.
	teardown := setupTestDB(t)
	defer teardown()

	// --- Pre-populate the database for the GET /posts test ---
	// We add a known post to the database to verify that the endpoint can retrieve it.
	stmt, err := db.Prepare("INSERT INTO posts(id, title, content) VALUES(?, ?, ?)")
	if err != nil {
		t.Fatalf("Failed to prepare statement: %v", err)
	}
	_, err = stmt.Exec("test-id-123", "Test Title", "Test Content")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	stmt.Close()
	// --- End of pre-population ---

	router := http.NewServeMux()
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/posts", handlePosts)

	t.Run("GET / returns welcome message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		expectedBody := "Welcome to the Bulletin Board API! Please use the /posts endpoint."
		if rec.Body.String() != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
		}
	})

	t.Run("GET /posts returns list of posts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/posts", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Check the JSON response.
		var posts []Post
		if err := json.Unmarshal(rec.Body.Bytes(), &posts); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		if len(posts) != 1 {
			t.Fatalf("expected 1 post, but got %d", len(posts))
		}
		if posts[0].Title != "Test Title" {
			t.Errorf("expected post title to be 'Test Title', but got %s", posts[0].Title)
		}
	})
}

// TestCreatePostEndpoint tests the POST /posts functionality.
func TestCreatePostEndpoint(t *testing.T) {
	// Setup a clean, empty database for this specific test.
	teardown := setupTestDB(t)
	defer teardown()

	router := http.NewServeMux()
	router.HandleFunc("/posts", handlePosts)

	postJSON := `{"title": "New Title", "content": "New Content"}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(postJSON))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// 1. Check the HTTP response
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	// 2. Verify the response body contains the created post
	var createdPost Post
	if err := json.Unmarshal(rec.Body.Bytes(), &createdPost); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	if createdPost.Title != "New Title" {
		t.Errorf("expected created post title to be 'New Title', but got %s", createdPost.Title)
	}
	if createdPost.ID == "" {
		t.Error("expected created post to have a non-empty ID")
	}

	// 3. Verify the data was actually written to the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE title = ?", "New Title").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database for new post: %v", err)
	}
	if count != 1 {
		t.Errorf("expected database to have 1 post with the new title, but found %d", count)
	}
}
