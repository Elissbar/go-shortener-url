package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/repository"
)

func main() {
	cfg := parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	myHandler := handler.MyHandler{Storage: &repository.MemoryStorage{}, Config: cfg, Logger: logger.Log.Sugar()}

	router := myHandler.Router()

	err := http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
