package main

import (
	"fmt"
	"net/http"

	"github.com/Rusich90/shurl.git/internal/handler"
)

func main() {
	http.HandleFunc("/", handler.CreateShortUrl)
	http.HandleFunc("/{id}", handler.GetOriginalUrl)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
