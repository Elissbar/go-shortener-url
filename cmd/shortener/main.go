package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/repository"
)

func main() {
	cfg := parseFlags()

	myHandler := handler.MyHandler{Storage: repository.MemoryStorage{}, Config: cfg}

	router := myHandler.Router()

	err := http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
