package filestorage

import (
	"strconv"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
	fileManager *FileManager
	serializer Serializer
}

func (ch *Cache) Save(token, url string) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.data[token] = url

	var records []model.URLRecord
	i := 1
	for shortURL, originalURL := range ch.data {
		records = append(
			records,
			model.URLRecord{
				UUID:        strconv.Itoa(i),
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			},
		)
		i++
	}
	byteData, err := ch.serializer.Marshal(records)
	err = ch.fileManager.SaveToFile(byteData)
	if err != nil {
		return err
	}

	return nil
}

func (ch *Cache) Get(token string) (string, bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	url, exists := ch.data[token]
	return url, exists
}
