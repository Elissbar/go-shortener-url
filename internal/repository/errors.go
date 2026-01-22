package repository

import "errors"

// Кастомные ошибки для обрабоки в хендлерах.
var (
	ErrURLExists      error = errors.New("URL already exists") // Ошибка сохранения существующего URL.
	ErrTokenIsDeleted error = errors.New("token is deleted") // Ошибка возникающая при попытке получения удаленного токена (ссылки).
	ErrTokenNotExist  error = errors.New("token is not exists") // Ошибка при попытке получения несуществующего токена (ссылки).
)
