package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
)

func main() {
	urls := make(map[string]string)
	myHandler := handler.MyHandler{Urls: urls}

	mux := http.NewServeMux()
	mux.HandleFunc("/", myHandler.CreateShortUrl)
	mux.HandleFunc("/{id}", myHandler.GetShortUrl)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
