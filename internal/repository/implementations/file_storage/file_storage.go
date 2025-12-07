package filestorage

import (
	"context"
	"strconv"
	"sync"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
)

type FileStorage struct {
	tokenURL    *sync.Map // token: url
	urlToken    *sync.Map // url: token
	fileManager *FileManager
	serializer  Serializer
}

func NewFileStorage(fm *FileManager, sr Serializer) (*FileStorage, error) {
	storage := &FileStorage{
		tokenURL:    &sync.Map{},
		urlToken:    &sync.Map{},
		fileManager: fm,
		serializer:  sr,
	}

	if err := fm.EnsureFile(); err != nil {
		return nil, err
	}

	if err := storage.loadFromFile(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (fs *FileStorage) loadFromFile() error {
	data, err := fs.fileManager.Read()
	if err != nil {
		return err
	}

	records, err := fs.serializer.Unmarshal(data)
	if err != nil {
		return err
	}

	for _, record := range records {
		fs.tokenURL.Store(record.ShortURL, record.OriginalURL)
		fs.urlToken.Store(record.OriginalURL, record.ShortURL)
	}
	return nil
}

func (fs *FileStorage) saveToFile() error {
	records := []model.URLRecord{}
	i := 0
	fs.tokenURL.Range(func(key, value any) bool {
		records = append(records, model.URLRecord{
			UUID:        strconv.Itoa(i),
			ShortURL:    key.(string),
			OriginalURL: value.(string),
		})

		i++
		return true
	})

	data, err := fs.serializer.Marshal(records)
	if err != nil {
		return err
	}

	return fs.fileManager.Write(data)

}

func (fs *FileStorage) Save(ctx context.Context, token, url, _, _ string) (string, error) {
	if val, ok := fs.urlToken.Load(url); ok {
		return val.(string), repository.ErrURLExists // Возвращаем токен, если URL уже существует
	}

	fs.tokenURL.Store(token, url)
	fs.urlToken.Store(url, token)

	if err := fs.saveToFile(); err != nil {
		return "", err
	}
	return token, nil
}

func (fs *FileStorage) SaveBatch(ctx context.Context, batch []model.ReqBatch, userID, baseURL string) error {
	for _, b := range batch {
		fs.Save(ctx, b.Token, b.OriginalURL, userID, baseURL)
	}
	return nil
}

func (fs *FileStorage) Get(ctx context.Context, token string) (string, error) {
	if val, ok := fs.tokenURL.Load(token); ok {
		return val.(string), nil // Возвращаем токен, если URL уже существует
	}
	return "", repository.ErrTokenNotExist
}

func (fs *FileStorage) GetAllUsersURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	return []model.URLRecord{}, nil
}

func (fs *FileStorage) DeleteByTokens(ctx context.Context, userID string, tokens []string) error {return nil}

func (fs *FileStorage) Close() error {
	return nil
}

func (fs *FileStorage) Ping() error {
	return nil
}
