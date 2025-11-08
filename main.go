package main

import (
	"encoding/json"
	"fmt"
	"html" // サニタイズのためにhtmlパッケージをインポート
	"log"
	"net/http"
	"sync" // 排他制御のためにsyncパッケージをインポート

	"github.com/google/uuid" // UUID生成のためにインポート
)

type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var (
	posts []Post
	mu    sync.RWMutex
)

func sanitize(s string) string {
	return html.EscapeString(s)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "掲示板APIへようこそ! /postsエンドポイントをご利用ください。")
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
		http.Error(w, `{"error": "許可されていないメソッドです"}`, http.StatusMethodNotAllowed)
	}
}

func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("投稿一覧のJSONエンコードに失敗しました: %v", err)
		http.Error(w, `{"error": "サーバー内部のエラー"}`, http.StatusInternalServerError)
	}
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var newPost Post

	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, `{"error": "リクエストボディの形式が不正です"}`, http.StatusBadRequest)
		return
	}

	newPost.ID = uuid.New().String()
	newPost.Title = sanitize((newPost.Title))
	newPost.Content = sanitize(newPost.Content)

	mu.Lock()
	posts = append(posts, newPost)
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Printf("新規投稿のJSONエンコードに失敗しました: %v", err)
		http.Error(w, `{"error": "サーバー内部のエラー"}`, http.StatusInternalServerError)
	}
}

func main() {
	posts = []Post{
		{"00000000-0000-0000-0000-000000000001", "テスト投稿1", "これは最初の投稿です。"},
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/posts", handlePosts)

	port := ":8080"
	fmt.Printf("サーバーをポート%sで起動します…\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
