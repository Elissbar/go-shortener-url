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
	log := logger.Log.Sugar()

	// Выбираем хранилище в зависимости от конфигурации
    var storage repository.Storage
    if cfg.FileStoragePath != "" {
        var err error
        storage, err = repository.NewFileStorage(cfg.FileStoragePath)
        if err != nil {
            log.Fatal("Failed to create file storage:", err)
        }
        defer storage.(*repository.FileStorage).Close()
    } else {
        storage = &repository.MemoryStorage{}
    }

	myHandler := &handler.MyHandler{
		Storage: storage, 
		Config: cfg, 
		Logger: log,
	}

	router := myHandler.Router()

	err := http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
