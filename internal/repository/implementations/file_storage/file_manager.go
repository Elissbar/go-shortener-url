package filestorage

import (
	"errors"
	"os"
	"path/filepath"
)

type FileManager struct {
	FilePath string
}

func NewFileManager(FilePath string) *FileManager {
	return &FileManager{FilePath: FilePath}
}

func (fm *FileManager) Read() ([]byte, error) {
	data, err := os.ReadFile(fm.FilePath)
	if errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil // возвращаем пустые данные вместо ошибки
	}
	return data, err
}

func (fm *FileManager) Write(data []byte) error {
	tempPath := fm.FilePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, fm.FilePath)
}

func (fm *FileManager) EnsureFile() error {
	// Создаем директорию
	dir := filepath.Dir(fm.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Если файла нет - создаем пустой
	if _, err := os.Stat(fm.FilePath); os.IsNotExist(err) {
		return os.WriteFile(fm.FilePath, []byte("[]"), 0644)
	}
	
	return nil
}

