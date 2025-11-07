package filestorage

import (
	"os"
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

func (fs *FileStorage) Save(token, url string) error {
	return fs.cache.Save(token, url)
}

func (fs *FileStorage) Get(token string) (string, bool) {
	return fs.cache.Get(token)
}
