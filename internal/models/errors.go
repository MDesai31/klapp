package models

import "errors"

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidPIN         = errors.New("models: invalid PIN")
	ErrAlreadyOpen        = errors.New("models: worker already has an open punch")
	ErrInvalidCredentials = errors.New("models: invalid username or password")
	ErrDuplicatePIN       = errors.New("models: PIN already in use by another worker")
	ErrDailyLimitExceeded = errors.New("models: worker reached daily punch-in limit")
	ErrEndBeforeStart     = errors.New("models: punch end time is not after its start time")
)
