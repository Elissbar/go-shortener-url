package filestorage

import (
	"context"
	"os"
	"strconv"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
)

type FileStorage struct {
	cache       *Cache
	fileManager *FileManager
	serializer  Serializer
}

func NewFileStorage(fm *FileManager, sr Serializer) (*FileStorage, error) {
	fs := &FileStorage{
		fileManager: fm,
		serializer:  sr,
		cache: &Cache{
			data: make(map[string]string),
		},
	}
	if err := fs.fileManager.EnsureFile(); err != nil {
		return nil, err
	}
	byteData, err := fs.fileManager.LoadFromFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	records, err := fs.serializer.Unmarshal(byteData)
	if err != nil {
		return nil, err
	}
	// Загружаем в память
	for _, record := range records {
		fs.cache.SaveToMemory(record.ShortURL, record.OriginalURL)
	}
	return fs, nil
}

func (fs *FileStorage) Save(ctx context.Context, token, url string) (string, error) {
	fs.cache.mu.Lock()
	defer fs.cache.mu.Unlock()

	for oldToken, val := range fs.cache.data {
		if val == url {
			return oldToken, repository.ErrURLExists
		}
	}
	fs.cache.data[token] = url
	return token, nil
}

func (fs *FileStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch) error {
	for _, b := range batch {
		fs.Save(ctx, b.Token, b.OriginalURL)
	}
	return nil
}

func (fs *FileStorage) Get(ctx context.Context, token string) (string, bool) {
	fs.cache.mu.RLock()
	defer fs.cache.mu.RUnlock()
	url, exists := fs.cache.data[token]
	return url, exists
}

func (fs *FileStorage) Close() error {
	var records []model.URLRecord
	i := 1
	for shortURL, originalURL := range fs.cache.data {
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
	byteData, err := fs.serializer.Marshal(records)
	if err != nil {
		return err
	}
	err = fs.fileManager.SaveToFile(byteData)
	if err != nil {
		return err
	}
	return nil
}
