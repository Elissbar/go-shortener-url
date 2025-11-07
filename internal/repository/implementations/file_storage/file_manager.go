package filestorage

import (
	"os"
)

type FileManager struct {
	FilePath string
}

func (fm *FileManager) LoadFromFile() ([]byte, error) {
	data, err := os.ReadFile(fm.FilePath)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (fm *FileManager) SaveToFile(data []byte) error {
	err := os.WriteFile(fm.FilePath, data, 0666)
	if err != nil {
		return err
	}
	return nil
}
