package repository

import "errors"

var (
	ErrURLExists      error = errors.New("URL already exists")
	ErrTokenIsDeleted error = errors.New("token is deleted")
	ErrTokenNotExist  error = errors.New("token is not exists")
)
