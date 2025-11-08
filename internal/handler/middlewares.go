package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
	"time"
)

// Кастомный ResponseWriter для gzip
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем уже сжатые форматы
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") ||
			strings.Contains(r.Header.Get("Content-Encoding"), "deflate") {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем поддержку gzip клиентом
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw := &gzipResponseWriter{
				Writer:             gzip.NewWriter(w),
				ResponseWriter: w,
			}
			defer gw.Writer.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gw, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Кастомный ResponseWriter для gzip
type gzipResponseWriter struct {
	Writer *gzip.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func ungzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, сжат ли запрос
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()

			r.Body = gz
		}

		next.ServeHTTP(w, r)
	})
}

func (h *MyHandler) LoggingMiddleware(handler http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		uri := r.RequestURI
		metnod := r.Method

		lw := responseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		handler.ServeHTTP(&lw, r)

		duration := time.Since(startTime)

		h.Logger.Infow("Request/Response data: ",
			"uri", uri,
			"method", metnod,
			"status", lw.responseData.status,
			"duration", int(duration),
			"size", lw.responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}
