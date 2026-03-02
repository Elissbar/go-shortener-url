package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "os"
	"reflect"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/observer"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
	"github.com/Elissbar/go-shortener-url/internal/service"
	"golang.org/x/sync/errgroup"
	// _ "net/http/pprof"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
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

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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
	go srvc.ProcessDeletions(shutdownCtx)

	httpServer := &http.Server{
		Addr: cfg.ServerURL,
		BaseContext: func(_ net.Listener) context.Context {
			return shutdownCtx
		},
		Handler: handler.NewHandler(srvc).Router(),
	}

	g, gCtx := errgroup.WithContext(shutdownCtx)
	g.Go(func() error {
		if cfg.EnableHTTPS != nil && (*cfg.EnableHTTPS) {
			log.Infof("🚀 HTTPS mode. Server started on %s", cfg.ServerURL)
			// err = http.ListenAndServeTLS(srvc.Config.ServerURL, "cert.pem", "key.pem", myHandler.Router())
			if err := httpServer.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("server error: %w", err)
			}
		} else {
			log.Infof("🚀 HTTP mode. Server started on %s", cfg.ServerURL)
			// err = http.ListenAndServe(srvc.Config.ServerURL, myHandler.Router())
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("server error: %w", err)
			}
		}
		return nil
	})
	g.Go(func() error {
		<-gCtx.Done()
		srvc.Logger.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		close(srvc.DeleteCh)

		log.Info("Server stopped")
		return nil
	})

	if err := g.Wait(); err != nil {
		srvc.Logger.Errorf("Application error: %v", err)
		os.Exit(1)
	}

	log.Info("Application stopped")
}
