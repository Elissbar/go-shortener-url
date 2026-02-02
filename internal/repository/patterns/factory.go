package patterns

import (
	"go.uber.org/zap"

	"github.com/Elissbar/go-shortener-url/internal/repository"
	databasestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/database_storage"
	filestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/file_storage"
	memorystorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/memory_storage"
)

func NewStorage(log *zap.SugaredLogger, databaseAdr, fileStoragePath string) (repository.Storage, error) {
	// Выбираем хранилище в зависимости от конфигурации
	if databaseAdr != "" {
		return databasestorage.NewDatabaseStorage(databaseAdr, log)
	}
	if fileStoragePath != "" {
		return filestorage.NewFileStorage(
			&filestorage.FileManager{FilePath: fileStoragePath},
			&filestorage.JSONSerializer{},
		)
	}

	return memorystorage.NewMemoryStorage()
}
