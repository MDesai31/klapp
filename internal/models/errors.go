package models

import "errors"

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidPIN         = errors.New("models: invalid PIN")
	ErrAlreadyOpen        = errors.New("models: worker already has an open punch")
	ErrInvalidCredentials = errors.New("models: invalid username or password")
	ErrDuplicatePIN       = errors.New("models: PIN already in use by another worker")
)
