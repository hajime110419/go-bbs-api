package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/posts", handlePosts)

	return router
}

func setupTestEnvironment() {
	mu.Lock()

	posts = []Post{
		{"00000000-0000-0000-0000-000000000001", "テスト投稿1", "これは最初の投稿です。"},
	}
	mu.Unlock()
}

func tearDownTestEnvironment() {
	mu.Lock()
	posts = nil
	mu.Unlock()
}

// Testという単語を1つにする
func TestRouting(t *testing.T) {
	setupTestEnvironment()
	defer tearDownTestEnvironment()
	router := setupRouter()

	// --- Test Case 1: GET / ---
	req1 := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()

	router.ServeHTTP(rec1, req1)

	//ステータスコードのチェック
	if rec1.Code != http.StatusOK {
		t.Errorf("GET /: expected status code %d, but got %d", http.StatusOK, rec1.Code)
	}

	//レスポンスボディのチェック
	expectedBody1 := "掲示板APIへようこそ! /postsエンドポイントをご利用ください。"
	if rec1.Body.String() != expectedBody1 {
		// 変数名を rec -> rec1 に修正
		t.Errorf("GET /: expected body %q, but got %q", expectedBody1, rec1.Body.String())
	}

	// --- Test Case : GET /posts ---
	req2 := httptest.NewRequest("GET", "/posts", nil)
	rec2 := httptest.NewRecorder() // req1, rec1 とは別の変数名を使う

	router.ServeHTTP(rec2, req2)

	//ステータスコードのチェック
	if rec2.Code != http.StatusOK {
		t.Errorf("GET /posts: expected status code %d, but got %d", http.StatusOK, rec2.Code)
	}

	//レスポンスボディのチェック
	expectedBody2 := `[{"id":"00000000-0000-0000-0000-000000000001","title":"テスト投稿1","content":"これは最初の投稿です。"}]` + "\n"
	if rec2.Body.String() != expectedBody2 {
		t.Errorf("GET /posts: expected body %q, but got %q", expectedBody2, rec2.Body.String())
	}
}
