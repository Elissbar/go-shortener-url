package observer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type HTTPSubscriber struct {
	ID  string
	URL string
}

func (h *HTTPSubscriber) Update(message model.AuditRequest) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshal data: %w", err)
	}
	fmt.Println("Data parsed")

	client := http.Client{}
	req, err := http.NewRequest("POST", h.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error create new request: %w", err)
	}
	fmt.Println("Client created")

	resp, err := client.Do(req)
	resp.Body.Close()
	return err
}

func (h *HTTPSubscriber) GetID() string {
	return h.ID
}
