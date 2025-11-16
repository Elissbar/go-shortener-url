package repository

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Storage interface {
	Save(ctx context.Context, token, url string) (string, error)
	SaveBatch(ctx context.Context, batch []model.ReqBatch) error
	Get(ctx context.Context, token string) (string, bool)
	Close() error
}
