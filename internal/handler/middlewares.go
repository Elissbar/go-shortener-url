package handler

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

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
				Writer:         gzip.NewWriter(w),
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

		h.Service.Logger.Infow("Request/Response data: ",
			"uri", uri,
			"method", metnod,
			"status", lw.responseData.status,
			"duration", int(duration),
			"size", lw.responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func (h *MyHandler) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		cookie, err := r.Cookie("user_id")

		if err != nil || cookie.Value == "" {
			cookie, userIDStr, err := generateAuthToken(h.Service.Config.JWTSecret)
			if err != nil {
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}
			userID = userIDStr
			http.SetCookie(w, cookie)
		} else {
			userID, err = verifyAuthToken(cookie.Value, h.Service.Config.JWTSecret)
			if err != nil {
				fmt.Println("Ошибка при проверке валидации куки")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateAuthToken(jwtSecret string) (*http.Cookie, string, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return &http.Cookie{}, "", err
	}
	userIDStr := userID.String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: userIDStr})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return &http.Cookie{}, "", err
	}

	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    tokenString,
		HttpOnly: true,
	}
	return cookie, userIDStr, nil
}

func verifyAuthToken(tokenString, jwtSecret string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims.UserID == "" {
		return "", fmt.Errorf("user_id is empty")
	}

	return claims.UserID, nil
}
