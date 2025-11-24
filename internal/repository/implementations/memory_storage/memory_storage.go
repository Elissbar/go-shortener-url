package memorystorage

import (
	"context"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
)

type MemoryStorage struct {
	TokenURL *sync.Map // token: url
	URLToken *sync.Map // url: token
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{
		TokenURL: &sync.Map{},
		URLToken: &sync.Map{},
	}, nil
}

func (ms *MemoryStorage) Save(ctx context.Context, token, url, userID string) (string, error) {
	if val, ok := ms.URLToken.Load(url); ok {
		return val.(string), repository.ErrURLExists // Возвращаем токен, если URL уже существует
	}

	ms.TokenURL.Store(token, url)
	ms.URLToken.Store(url, token)
	return token, nil
}

func (ms *MemoryStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch, userID string) error {
	for _, b := range batch {
		ms.Save(ctx, b.Token, b.OriginalURL, userID)
	}
	return nil
}

func (ms *MemoryStorage) Get(ctx context.Context, token string) (string, bool) {
	if val, ok := ms.TokenURL.Load(token); ok {
		return val.(string), true // Возвращаем токен, если URL уже существует
	}
	return "", false
}

func (ms *MemoryStorage) GetAllUsersURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	return []model.URLRecord{}, nil
}

func (ms *MemoryStorage) Close() error { return nil }

func (ms *MemoryStorage) Ping() error { return nil }
