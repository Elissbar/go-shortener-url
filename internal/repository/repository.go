package repository

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Storage interface {
	Save(ctx context.Context, token, url, userID, baseURL string) (string, error)
	SaveBatch(ctx context.Context, batch []model.ReqBatch, userID, baseURL string) error
	Get(ctx context.Context, token string) (string, bool)
	GetAllUsersURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
	Close() error
	Ping() error
}
