package repository

import "sync"

type MemoryStorage struct {
	Urls sync.Map
}

func (ms *MemoryStorage) Save(token, url string) error {
	ms.Urls.Store(token, url)
	return nil
}

func (ms *MemoryStorage) Get(token string) (string, bool) {
	val, ok := ms.Urls.Load(token)
	if !ok {
		return "", false
	}
	return val.(string), true
}
