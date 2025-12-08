package main

import (
	"net/http"
	"reflect"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
	"github.com/Elissbar/go-shortener-url/internal/service"
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

	srvc := service.NewService(log, storage)
	go srvc.ProcessDeletions()
	myHandler := handler.NewHandler(storage, cfg, log, srvc)

	router := myHandler.Router()

	err = http.ListenAndServe(cfg.ServerURL, router)
	if err != nil {
		panic(err)
	}
}
