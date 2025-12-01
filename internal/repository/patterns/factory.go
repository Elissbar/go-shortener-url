package patterns

import (
	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	databasestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/database_storage"
	filestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/file_storage"
	memorystorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/memory_storage"
	"go.uber.org/zap"
)

func NewStorage(cfg *config.Config, log *zap.SugaredLogger) (repository.Storage, error) {
	// Выбираем хранилище в зависимости от конфигурации
	if cfg.DatabaseAdr != "" {
		return databasestorage.NewDatabaseStorage(cfg.DatabaseAdr, log)
	}
	if cfg.FileStoragePath != "" {
		return filestorage.NewFileStorage(
			&filestorage.FileManager{FilePath: cfg.FileStoragePath},
			&filestorage.JSONSerializer{},
		)
	}

	return memorystorage.NewMemoryStorage()
}
