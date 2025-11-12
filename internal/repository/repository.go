package repository

import "context"

type Storage interface {
	Save(ctx context.Context, token, url string) error
	Get(ctx context.Context, token string) (string, bool)
	Close() error
}
