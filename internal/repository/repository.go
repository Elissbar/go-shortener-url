package repository

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Storage interface {
	Reader
	Writer
	Manager
}

type Reader interface {
	Get(ctx context.Context, token string) (string, error)
	GetAllUsersURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
}

type Writer interface {
	Save(ctx context.Context, token, url, userID, baseURL string) (string, error)
	SaveBatch(ctx context.Context, batch []model.ReqBatch, userID, baseURL string) error
	DeleteByTokens(ctx context.Context, userID string, tokens []string) error
}

type Manager interface {
	Close() error
	Ping() error
}
