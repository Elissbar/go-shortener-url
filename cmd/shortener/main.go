package main

import (
	"net/http"
	"reflect"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		panic(err)
	}

	log, err := logger.NewSugaredLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	storage, err := patterns.NewStorage(cfg, log)
	if err != nil {
		panic(err)
	}
	defer storage.Close()
	log.Infow("Storage type:",
		"type", reflect.TypeOf(storage),
	)

	// myHandler := &handler.MyHandler{
	// 	Storage: storage,
	// 	Config:  cfg,
	// 	Logger:  log,
	// 	DeleteCh: make(chan []string, 1000),
	// }
	myHandler := handler.NewService(storage, cfg, log)

	router := myHandler.Router()

	err = http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
