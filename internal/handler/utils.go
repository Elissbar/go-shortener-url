package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
)

func generateToken() (string, error) {
	// Генерируем токен - id короткой ссылки
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)
	token = token[:len(token)-1]
	return token, nil
}

func updateLinksStore(urls map[string]string) error {
	updatedData, err := json.Marshal(urls)
	if err != nil {
		return err
	}

	err = os.WriteFile("urls.json", updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetAllLinks(urls *map[string]string) error {
	jsonFile, err := os.Open("urls.json")
	defer jsonFile.Close()
	if err != nil {
		return err
	}

	val, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(val, &urls)
	if err != nil {
		return err
	}

	return nil
}
