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

func (db *DBStorage) Save(ctx context.Context, token, url, userID, baseURL string) (string, error) {
	_, err := db.DB.ExecContext(ctx, "INSERT INTO shorted_links (token, url, user_id, shorted_url) VALUES ($1, $2, $3, $4)", token, url, userID, baseURL+token)
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

func (db *DBStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch, userID, baseURL string) error {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, b := range batch {
		_, err := tx.ExecContext(ctx, "INSERT INTO shorted_links (token, url, user_id, shorted_url) VALUES ($1, $2, $3, $4)", b.Token, b.OriginalURL, userID, baseURL+b.Token)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (db *DBStorage) Get(ctx context.Context, token string) (string, error) {
	row := db.DB.QueryRowContext(ctx, "SELECT url, deleted FROM shorted_links WHERE token = $1", token)

	var url string
	var deleted bool
	err := row.Scan(&url, &deleted)

	if err != nil {
		return "", err
	}
	if deleted {
		return url, repository.ErrTokenIsDeleted
	}

	return url, nil
}

func (db *DBStorage) GetAllUsersURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	rows, err := db.DB.QueryContext(ctx, "SELECT shorted_url, url FROM shorted_links WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	rows.Err()
	defer rows.Close()

	records := []model.URLRecord{}
	for rows.Next() {
		var record model.URLRecord
		err := rows.Scan(
			&record.ShortURL,
			&record.OriginalURL,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (db *DBStorage) DeleteByTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
        return nil
    }
    
    // Batch update - один запрос для всех токенов
    query := "UPDATE shorted_links SET deleted = true WHERE token = ANY($1)"
    _, err := db.DB.ExecContext(ctx, query, pq.Array(tokens))
    return err
}

func (db *DBStorage) Close() error {
	return db.DB.Close()
}

func (db *DBStorage) Ping() error {
	return db.DB.Ping()
}
