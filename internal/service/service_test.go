package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/Elissbar/go-shortener-url/internal/service"
)

func TestGenerateToken(t *testing.T) {
	s := &service.Service{}

	t.Run("zero size", func(t *testing.T) {
		token, err := s.GenerateToken(0)
		if err != nil {
			t.Fatalf("unexpected error for size 0: %v", err)
		}
		if token != "" {
			t.Errorf("expected empty token for size 0, got %q", token)
		}
	})

	t.Run("randomness", func(t *testing.T) {
		// Проверяем, что токены разные при разных вызовах
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := s.GenerateToken(8)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tokens[token] {
				t.Errorf("duplicate token generated: %s", token)
			}
			tokens[token] = true
		}
	})

}

func TestGetToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success on first attempt", func(t *testing.T) {
		mockStorage := &mockStorage{}
		s := &service.Service{Storage: mockStorage}

		token, err := s.GetToken(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("storage error", func(t *testing.T) {
		expectedErr := errors.New("storage failure")
		mockStorage := &mockStorage{
			getFunc: func(ctx context.Context, token string) (string, error) {
				return "", expectedErr
			},
		}
		s := &service.Service{Storage: mockStorage}

		_, err := s.GetToken(ctx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected storage error, got %v", err)
		}
	})
}

// Мок хранилища
type mockStorage struct {
	repository.Storage
	getFunc func(ctx context.Context, token string) (string, error)
}

func (m *mockStorage) Get(ctx context.Context, token string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, token)
	}
	return "", repository.ErrTokenNotExist
}
