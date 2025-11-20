package implementations

import (
	"context"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type MemoryStorage struct {
	TokenURL sync.Map // token: url
	URLToken sync.Map // url: token
}

func (ms *MemoryStorage) Save(ctx context.Context, token, url string) (string, error) {
	if val, ok := ms.URLToken.Load(url); ok {
		return val.(string), nil // Возвращаем токен, если URL уже существует
	}

	ms.TokenURL.Store(token, url)
	ms.URLToken.Store(url, token)
	return token, nil
}

func (ms *MemoryStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch) error {
	for _, b := range batch {
		ms.Save(ctx, b.Token, b.OriginalURL)
	}
	return nil
}

func (ms *MemoryStorage) Get(ctx context.Context, token string) (string, bool) {
	val, ok := ms.TokenURL.Load(token)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func (ms *MemoryStorage) Close() error { return nil }

func (ms *MemoryStorage) Ping() error { return nil }
