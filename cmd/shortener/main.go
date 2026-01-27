package main

import (
	"net/http"
	"reflect"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/observer"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
	"github.com/Elissbar/go-shortener-url/internal/service"
	// _ "net/http/pprof"
)

// @title Shortener URL API
// @host localhost:8080
// @schemes http
// @BasePath /
func main() {
	// для запуска pprof на отдельном порту
	// go func() {
	//     http.ListenAndServe("localhost:6060", nil)
	// }()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	log, err := logger.NewSugaredLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	storage, err := patterns.NewStorage(log, cfg.DatabaseAdr, cfg.FileStoragePath)
	if err != nil {
		panic(err)
	}
	log.Infow("Storage type:", "type", reflect.TypeOf(storage))

	event := observer.NewEvent()
	if cfg.AuditFile != "" {
		event.Subscribe(&observer.FileSubscriber{ID: "FileSub", FilePath: cfg.AuditFile})
		log.Infow("Registered file audit. Audit file: " + cfg.AuditFile)
	}
	if cfg.AuditURL != "" {
		event.Subscribe(&observer.HTTPSubscriber{ID: "HTTPSub", URL: cfg.AuditURL})
		log.Infow("Registered http auditt. URL for audit: " + cfg.AuditURL)
	}

	srvc := service.NewService(cfg, log, storage, event)
	defer srvc.Helper.Close()
	go srvc.ProcessDeletions()

	myHandler := handler.NewHandler(srvc)
	err = http.ListenAndServe(srvc.Config.ServerURL, myHandler.Router())
	if err != nil {
		panic(err)
	}
}
