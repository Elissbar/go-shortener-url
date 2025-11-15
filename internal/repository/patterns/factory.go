package patterns

import (
	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/Elissbar/go-shortener-url/internal/repository/implementations"
	databasestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/database_storage"
	filestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/file_storage"
)

func NewStorage(cfg *config.Config) (repository.Storage, error) {
	// Выбираем хранилище в зависимости от конфигурации
	if cfg.DatabaseAdr != "" {
		return databasestorage.NewDatabaseStorage(cfg.DatabaseAdr)
	}
	if cfg.FileStoragePath != "" {
		return filestorage.NewFileStorage(
			&filestorage.FileManager{FilePath: cfg.FileStoragePath},
			&filestorage.JSONSerializer{},
		)
	}

	return &implementations.MemoryStorage{}, nil
}
