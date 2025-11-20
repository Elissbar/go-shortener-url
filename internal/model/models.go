package model

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ReqBatch struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
	Token       string
}

type RespBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url,omitempty"`
}
