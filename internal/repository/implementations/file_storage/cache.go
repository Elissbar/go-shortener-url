package filestorage

import (
	"sync"
)

type Cache struct {
	mu          sync.RWMutex
	data        map[string]string
	fileManager *FileManager
	serializer  Serializer
}

func (ch *Cache) Save(token, url string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.data[token] = url
}

func (ch *Cache) Get(token string) (string, bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	url, exists := ch.data[token]
	return url, exists
}
