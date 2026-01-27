package observer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type FileSubscriber struct {
	ID       string
	FilePath string
}

func (fs *FileSubscriber) Update(message model.AuditRequest) error {
	f, err := os.OpenFile(fs.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file for subscriber: %w", err)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	if _, err := f.WriteString(string(data)); err != nil {
		f.Close()
		return fmt.Errorf("error writing file for subscriber: %w", err)
	}
	return nil
}

func (fs *FileSubscriber) GetID() string {
	return fs.ID
}
