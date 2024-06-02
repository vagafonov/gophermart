package errs

import (
	"errors"
)

var ErrJWTTokenNotValid = errors.New("JWT token is not valid")
