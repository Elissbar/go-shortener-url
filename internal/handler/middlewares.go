package handler

import (
	"net/http"
	"time"
)

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