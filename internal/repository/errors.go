package repository

import "errors"

var (
	ErrURLExists error = errors.New("URL already exists")
)
