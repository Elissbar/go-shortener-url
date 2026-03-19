package handler

import (
	"encoding/json"
	"fmt"
)

func generateBatchData(n int) []byte {
	type request struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	batch := make([]request, 0, n)
	for i := 1; i <= n; i++ {
		batch = append(batch, request{
			CorrelationID: fmt.Sprintf("corr_%06d", i),
			OriginalURL:   fmt.Sprintf("https://example.com/path/%d?param=value%d&ref=test", i, i),
		})
	}

	data, err := json.Marshal(batch)
	if err != nil {
		// В реальном коде обработайте ошибку
		return []byte("[]")
	}
	return data
}
