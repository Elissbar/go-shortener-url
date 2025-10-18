package repository

import (
	"sync"
)

type Storage interface {
	Save(token, url string)
	Get(token string) (string, error)
}

type MemoryStorage struct {
	Urls sync.Map
}

func (ms *MemoryStorage) Save(token, url string) {
	ms.Urls.Store(token, url)
}

func (ms *MemoryStorage) Get(token string) (string, bool) {
	val, ok := ms.Urls.Load(token)
	if !ok {
		return "", false
	}
	return val.(string), true
}
