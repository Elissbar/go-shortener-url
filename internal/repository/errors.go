package repository

import "errors"

// Кастомные ошибки для обрабоки в хендлерах.
var (
	// Ошибка сохранения существующего URL.
	ErrURLExists error = errors.New("URL already exists")
	// Ошибка возникающая при попытке получения удаленного токена (ссылки).
	ErrTokenIsDeleted error = errors.New("token is deleted")
	// Ошибка при попытке получения несуществующего токена (ссылки).
	ErrTokenNotExist error = errors.New("token is not exists")
	// Сокращенные ссылки пользователя не найдены.
	ErrUserHasNoURL error = errors.New("user hasn't URLs")
	// Подсеть запрещена
	ErrSubnetForbidden error = errors.New("subnet forbidden")
)
