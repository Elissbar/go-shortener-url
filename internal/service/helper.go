package service

import (
	"github.com/Elissbar/go-shortener-url/internal/repository"
	databasestorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/database_storage"
)

// Helper - вспомогательный тип для работы с хранилищем.
type Helper struct {
	Storage *repository.Storage
}

func (h *Helper) Ping() error {
	db, ok := (*h.Storage).(*databasestorage.DBStorage)
	if ok {
		return db.DB.Ping()
	}
	return nil
}

func (h *Helper) Close() error {
	db, ok := (*h.Storage).(*databasestorage.DBStorage)
	if ok {
		return db.DB.Close()
	}
	return nil
}
