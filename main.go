package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "掲示板APIへようこそ!")
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, "ここは投稿一覧を返す場所です")
	case "POST":
		fmt.Fprint(w, "ここは新しい投稿を作成する場所です")
	default:
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/posts", handlePosts)

	port := ":8080"
	fmt.Printf("サーバーをポート%sで起動します…\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
