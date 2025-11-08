package filestorage

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileManager struct {
	FilePath string
}

func (fm *FileManager) LoadFromFile() ([]byte, error) {
	if _, err := os.Stat(fm.FilePath); os.IsNotExist(err) {
		return []byte{}, nil
	}

	data, err := os.ReadFile(fm.FilePath)
	fmt.Println(fm.FilePath)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (fm *FileManager) SaveToFile(data []byte) error {
	dir := filepath.Dir(fm.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	err := os.WriteFile(fm.FilePath, data, 0666)
	if err != nil {
		return err
	}
	return nil
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
