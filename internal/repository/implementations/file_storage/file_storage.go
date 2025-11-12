package filestorage

import (
	"context"
	"os"
	"strconv"

	"github.com/Elissbar/go-shortener-url/internal/model"
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
			data:        make(map[string]string),
			fileManager: fm,
			serializer:  sr,
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
		fs.cache.Save(record.ShortURL, record.OriginalURL)
	}
	return fs, nil
}

func (fs *FileStorage) Save(ctx context.Context, token, url string) error {
	fs.cache.Save(token, url)
	return nil
}

func (fs *FileStorage) Get(ctx context.Context, token string) (string, bool) {
	return fs.cache.Get(token)
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
	err = fs.fileManager.SaveToFile(byteData)
	if err != nil {
		return err
	}
	return nil
}
