package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/observer"
)

func audit(event *observer.Event, action, userID, url string) {
	event.Update(model.AuditRequest{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	})
}

func getFullBaseURL(baseURL string) string {
	fullURL := baseURL
	if !strings.HasSuffix(baseURL, "/") {
		fullURL = baseURL + "/"
	}
	return fullURL
}

func prepareHandler(r *http.Request) (string, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		return "", ctx, cancel, fmt.Errorf("Error get user id")
	}

	return userID, ctx, cancel, nil
}
