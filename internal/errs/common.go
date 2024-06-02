package errs

import (
	"errors"
)

var (
	ErrAlreadyExists              = errors.New("already exists")
	ErrNotFound                   = errors.New("not found")
	ErrInsufficientFundsOnBalance = errors.New("insufficient funds on balance")
)
