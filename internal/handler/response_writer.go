package handler

import (
	"net/http"
)

type responseData struct {
	status, size int
}

type responseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseData.size += size
	return size, err
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.ResponseWriter.WriteHeader(status)
	rw.responseData.status = status
}
