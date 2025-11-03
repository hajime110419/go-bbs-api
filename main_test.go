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

// Testという単語を1つにする
func TestRouting(t *testing.T) {
	router := setupRouter()

	// --- Test Case 1: GET / ---
	req1 := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()

	router.ServeHTTP(rec1, req1)

	// 検証①: ステータスコードのチェック
	if rec1.Code != http.StatusOK {
		t.Errorf("GET /: expected status code %d, but got %d", http.StatusOK, rec1.Code)
	}

	// 検証②: レスポンスボディのチェック
	expectedBody1 := "掲示板APIへようこそ!" // 元のコードで「！」が抜けていたので追加
	if rec1.Body.String() != expectedBody1 {
		// 変数名を rec -> rec1 に修正
		t.Errorf("GET /: expected body %q, but got %q", expectedBody1, rec1.Body.String())
	}

	// --- Test Case 2: GET /posts ---
	req2 := httptest.NewRequest("GET", "/posts", nil)
	rec2 := httptest.NewRecorder() // req1, rec1 とは別の変数名を使う

	router.ServeHTTP(rec2, req2)

	// 検証①: ステータスコードのチェック
	if rec2.Code != http.StatusOK {
		t.Errorf("GET /posts: expected status code %d, but got %d", http.StatusOK, rec2.Code)
	}

	// 検証②: レスポンスボディのチェック
	expectedBody2 := "ここは投稿一覧を返す場所です"
	if rec2.Body.String() != expectedBody2 {
		t.Errorf("GET /posts: expected body %q, but got %q", expectedBody2, rec2.Body.String())
	}
}
