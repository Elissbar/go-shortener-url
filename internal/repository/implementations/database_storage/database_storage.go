package databasestorage

import (
	"context"
	"database/sql"
	"errors"

	"fmt"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgerrcode"

	"github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DBStorage struct {
	DB             *sql.DB
	connectionData string
}

func NewDatabaseStorage(connectionData string) (*DBStorage, error) {
	db, err := sql.Open("postgres", connectionData)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// ПЕРВОЕ: проверяем соединение с БД
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &DBStorage{DB: db, connectionData: connectionData}

	// ВТОРОЕ: применяем миграции
	if err := storage.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return storage, nil
}

func (db *DBStorage) Migrate() error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (db *DBStorage) Save(ctx context.Context, token, url string) (string, error) {
	_, err := db.DB.ExecContext(ctx, "INSERT INTO shorted_links (token, url) VALUES ($1, $2)", token, url)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var oldToken string
			row := db.DB.QueryRowContext(ctx, "SELECT token FROM shorted_links WHERE url = $1", url)
			err = row.Scan(&oldToken)
			if err != nil {
				return "", err
			}
			return oldToken, repository.ErrURLExists
		}
		return "", err
	}
	return token, nil
}

func (db *DBStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch) error {
	stmt, err := db.DB.PrepareContext(ctx, "INSERT INTO shorted_links (token, url) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range batch {
		_, err := stmt.ExecContext(ctx, b.Token, b.OriginalURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DBStorage) Get(ctx context.Context, token string) (string, bool) {
	row := db.DB.QueryRowContext(ctx, "SELECT url FROM shorted_links WHERE token = $1", token)

	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", false
	}
	return value, true
}

func (db *DBStorage) Close() error {
	return db.DB.Close()
}
