package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
)

func main() {
	cfg := parseFlags()

	log, err := logger.NewSugaredLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	storage, err := patterns.NewStorage(cfg)
	if err != nil {
		panic(err)
	}

	myHandler := &handler.MyHandler{
		Storage: storage,
		Config:  cfg,
		Logger:  log,
	}

	router := myHandler.Router()

	err = http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
