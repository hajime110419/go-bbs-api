package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "掲示板APIへようこそ!")
}

func main() {
	http.HandleFunc("/", handleRoot)

	port := ":8080"
	fmt.Printf("サーバーをポート%sで起動します…\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
