package filestorage

import (
	"sync"
)

type Cache struct {
	mu          sync.RWMutex
	data        map[string]string
}

func (ch *Cache) SaveToMemory(token, url string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.data[token] = url
}
