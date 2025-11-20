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

func (fm *FileManager) ensureFile() error {
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

func (fm *FileManager) Read() ([]byte, error) {
	if err := fm.ensureFile(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fm.FilePath)
	if errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil // ← возвращаем пустые данные вместо ошибки
	}
	return data, err
}

func (fm *FileManager) Write(data []byte) error {
	// dir := filepath.Dir(fm.FilePath)
	// if err := os.MkdirAll(dir, 0755); err != nil {
	// 	return err
	// }

	tempPath := fm.FilePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, fm.FilePath)
}

// package filestorage

// import (
// 	"os"
// 	"path/FilePath"
// )

// type FileManager struct {
// 	FilePath string
// }

// func (fm *FileManager) LoadFromFile() ([]byte, error) {
// 	if _, err := os.Stat(fm.FilePath); os.IsNotExist(err) {
// 		return []byte{}, nil
// 	}

// 	data, err := os.ReadFile(fm.FilePath)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	return data, nil
// }

// func (fm *FileManager) SaveToFile(data []byte) error {
// 	dir := FilePath.Dir(fm.FilePath)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return err
// 	}

// 	err := os.WriteFile(fm.FilePath, data, 0666)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (fm *FileManager) EnsureFile() error {
// 	// Создаем директорию
// 	dir := FilePath.Dir(fm.FilePath)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return err
// 	}

// 	// Если файла нет - создаем пустой
// 	if _, err := os.Stat(fm.FilePath); os.IsNotExist(err) {
// 		return os.WriteFile(fm.FilePath, []byte("[]"), 0644)
// 	}

// 	return nil
// }
