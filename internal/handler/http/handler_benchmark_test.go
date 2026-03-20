package handler

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/service"
)

var sink interface{}

func BenchmarkBatch(b *testing.B) {
	size := 100
	srvc := service.Service{
		Config: &config.Config{BaseURL: "http://localhost:8080/"},
	}
	shortedURL := srvc.Config.BaseURL + "123"

	var reqBatch []model.ReqBatch
	data := generateBatchData(size)
	if err := json.Unmarshal(data, &reqBatch); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.Run("pre-allocate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			respBatch := make([]model.RespBatch, 0, size)
			for i := range size {
				batch := &reqBatch[i]

				batch.Token = "123"
				respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
			}
			sink = respBatch
		}
	})

	b.Run("no alloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var respBatch []model.RespBatch
			for i := range size {
				batch := &reqBatch[i]

				batch.Token = "123"
				respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
			}
			sink = respBatch
		}
	})
}

func BenchmarkParseJson(b *testing.B) {
	size := 10000
	data := generateBatchData(size)

	b.ResetTimer()

	b.Run("create decoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var reqBatch []model.ReqBatch
			dec := json.NewDecoder(bytes.NewReader(data))
			if err := dec.Decode(&reqBatch); err != nil {
				b.Fatal("Create decoder error: " + err.Error())
			}
		}
	})

	b.Run("unmarshal", func(b *testing.B) {
		var reqBatch []model.ReqBatch
		err := json.Unmarshal(data, &reqBatch)
		if err != nil {
			b.Fatal("Unmarshal error: " + err.Error())
		}
	})
}
