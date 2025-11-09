package databasestorage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	DB *sql.DB
}

func NewDatabaseStorage(connectionData string) (*DBStorage, error) {
	data := strings.Split(connectionData, ":")
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", data[0], data[1], data[2], data[3])

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	storage := &DBStorage{DB: db}
	return storage, nil
}

func (db *DBStorage) Save(token, url string) error { return nil }

func (db *DBStorage) Get(token string) (string, bool) { return "", false }

func (db *DBStorage) Close() error {
	return db.DB.Close()
}
