package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
)

func main() {
	// Явно пишем в stderr для отладки
	fmt.Fprintln(os.Stderr, "=== SERVER STARTING ===")

	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		// os.Exit(1)
	}

	// Принудительно устанавливаем порт
	cfg.ServerURL = ":8080"
	fmt.Fprintf(os.Stderr, "Using address: %s\n", cfg.ServerURL)

	log, err := logger.NewSugaredLogger("debug") // максимальное логирование
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger error: %v\n", err)
		// os.Exit(1)
	}

	storage, err := patterns.NewStorage(cfg)
	if err != nil {
		log.Errorw("Storage error", "error", err)
		// os.Exit(1)
	}
	defer storage.Close()

	myHandler := &handler.MyHandler{
		Storage: storage,
		Config:  cfg,
		Logger:  log,
	}

	server := &http.Server{
		Addr:    cfg.ServerURL,
		Handler: myHandler.Router(),
	}

	go func() {
		log.Infow("Starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("Server failed", "error", err)
		}
	}()

	// Даем время запуститься
	// time.Sleep(100 * time.Millisecond)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	log.Info("Waiting for shutdown signals...")
	sig := <-quit

	log.Infow("Shutting down", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorw("Shutdown error", "error", err)
	}

	log.Info("Server stopped")
	fmt.Fprintln(os.Stderr, "=== SERVER STOPPED ===")
}


// package main

// import (
// 	"net/http"

// 	"github.com/Elissbar/go-shortener-url/internal/handler"
// 	"github.com/Elissbar/go-shortener-url/internal/logger"
// 	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
// )

// func main() {
// 	cfg, err := parseFlags()
// 	if err != nil {
// 		panic(err)
// 	}

// 	log, err := logger.NewSugaredLogger(cfg.LogLevel)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer log.Sync()

// 	storage, err := patterns.NewStorage(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer storage.Close()

// 	myHandler := &handler.MyHandler{
// 		Storage: storage,
// 		Config:  cfg,
// 		Logger:  log,
// 	}

// 	router := myHandler.Router()

// 	err = http.ListenAndServe(cfg.ServerURL, router)
// 	if err != nil {
// 		panic(err)
// 	}
// }