// package filestorage

// import (
// 	"encoding/json"

// 	"github.com/Elissbar/go-shortener-url/internal/model"
// )

// type Serializer interface {
// 	Marshal(records []model.URLRecord) ([]byte, error)
// 	Unmarshal(data []byte, records []model.URLRecord) error
// }

// type JSONSerializer struct{}

// func (js JSONSerializer) Marshal(records []model.URLRecord) ([]byte, error) {
// 	return json.MarshalIndent(records, "", "  ")
// }

// func (js JSONSerializer) Unmarshal(data []byte, records []model.URLRecord) error {
// 	return json.Unmarshal(data, &records)
// }

package filestorage

import (
	"encoding/json"

	"github.com/Elissbar/go-shortener-url/internal/model"
)

type Serializer interface {
	Marshal(records []model.URLRecord) ([]byte, error)
	Unmarshal([]byte) ([]model.URLRecord, error)
}

type JSONSerializer struct{}

func (js *JSONSerializer) Marshal(records []model.URLRecord) ([]byte, error) {
	marshaledData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return []byte{}, err
	}
	return marshaledData, nil
}

func (js *JSONSerializer) Unmarshal(data []byte) ([]model.URLRecord, error) {
	var records []model.URLRecord
	err := json.Unmarshal(data, &records)
	if err != nil {
		return []model.URLRecord{}, err
	}
	return records, nil
}
