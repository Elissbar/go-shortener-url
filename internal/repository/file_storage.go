package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type FileStorage struct {
	filePath string
	mu       sync.RWMutex
	data     map[string]string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	storage := &FileStorage{
		filePath: filePath,
		data:     make(map[string]string),
	}

	// Загружаем данные при старте
	if err := storage.loadFromFile(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (fs *FileStorage) Save(token, url string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Сохраняем в память
	fs.data[token] = url

	// Сохраняем на диск
	return fs.saveToFile()
}

func (fs *FileStorage) Get(token string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	url, exists := fs.data[token]
	return url, exists
}

func (fs *FileStorage) loadFromFile() error {
	// Проверяем существует ли файл
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return nil
	}

	// Читаем файл
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return err
	}

	// Парсим JSON
	var records []model.URLRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	// Загружаем в память
	for _, record := range records {
		fs.data[record.ShortURL] = record.OriginalURL
	}

	return nil
}

func (fs *FileStorage) saveToFile() error {
	// Преобразуем данные в JSON формат
	var records []model.URLRecord
	i := 1
	for shortURL, originalURL := range fs.data {
		records = append(records, model.URLRecord{
			UUID:        string(i),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
		i++
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	// Записываем в файл
	return os.WriteFile(fs.filePath, data, 0644)
}

func (fs *FileStorage) Close() error {
	return fs.saveToFile() // сохраняем при закрытии
}
