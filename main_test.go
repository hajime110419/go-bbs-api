package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupRouter initializes a new ServeMux and registers the application's handlers.
// It returns the configured router for use in tests.
func setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/posts", handlePosts)
	return router
}

// setupTestEnvironment prepares the global state for a test run.
// It initializes the 'posts' slice with a known set of data.
func setupTestEnvironment() {
	mu.Lock()
	defer mu.Unlock()
	posts = []Post{
		{"00000000-0000-0000-0000-000000000001", "Test Post 1", "This is the first post."},
	}
}

// tearDownTestEnvironment cleans up the global state after a test run.
// It resets the 'posts' slice to nil to ensure test isolation.
func tearDownTestEnvironment() {
	mu.Lock()
	defer mu.Unlock()
	posts = nil
}

// TestEndpoints uses subtests to test the application's HTTP endpoints.
func TestEndpoints(t *testing.T) {
	// Set up the environment once for all subtests in this function.
	setupTestEnvironment()
	defer tearDownTestEnvironment()

	router := setupRouter()

	// Define test cases as a table to keep the code DRY (Don't Repeat Yourself).
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET / returns the welcome message",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Welcome to the Bulletin Board API! Please use the /posts endpoint.",
		},
		{
			name:           "GET /posts returns the list of posts",
			method:         http.MethodGet,
			path:           "/posts",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"00000000-0000-0000-0000-000000000001","title":"Test Post 1","content":"This is the first post."}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Check the status code.
			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, rec.Code)
			}

			// Check the response body.
			// Using strings.TrimSpace to handle potential trailing newlines from json.Encoder.
			if strings.TrimSpace(rec.Body.String()) != tc.expectedBody {
				t.Errorf("expected body %q, but got %q", tc.expectedBody, rec.Body.String())
			}
		})
	}
}

// TestCreatePostEndpoint tests the POST /posts functionality.
func TestCreatePostEndpoint(t *testing.T) {
	// This test needs a clean slate, so we manage the environment inside it.
	tearDownTestEnvironment() // Clean up before test
	defer tearDownTestEnvironment()

	router := setupRouter()

	t.Run("POST /posts creates a new post", func(t *testing.T) {
		postJSON := `{"title": "New Title", "content": "New Content"}`
		req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(postJSON))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Check status code.
		if rec.Code != http.StatusCreated {
			t.Errorf("expected status code %d, but got %d", http.StatusCreated, rec.Code)
		}

		// Check if the post was actually added.
		mu.RLock()
		if len(posts) != 1 {
			t.Errorf("expected 1 post to be created, but found %d", len(posts))
		}
		if posts[0].Title != "New Title" {
			t.Errorf("expected new post title to be 'New Title', but got %s", posts[0].Title)
		}
		mu.RUnlock()
	})
}
