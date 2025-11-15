package implementations

import (
	"context"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type MemoryStorage struct {
	Urls sync.Map
}

func (ms *MemoryStorage) Save(ctx context.Context, token, url string) error {
	ms.Urls.Store(token, url)
	return nil
}

func (ms *MemoryStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch) error {
	for _, b := range batch {
		ms.Save(ctx, b.Token, b.OriginalURL)
	}
	return nil
}

func (ms *MemoryStorage) Get(ctx context.Context, token string) (string, bool) {
	val, ok := ms.Urls.Load(token)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func (ms *MemoryStorage) Close() error { return nil }
