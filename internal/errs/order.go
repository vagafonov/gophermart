package errs

import (
	"errors"
)

//nolint:lll
var (
	ErrOrderNumberHasAlreadyBeenUploadedByCurrentUser = errors.New("order number has already been uploaded by current user")
	ErrOrderNumberHasAlreadyBeenUploadedBySomeUser    = errors.New("order number has already been uploaded by some user")
	ErrInvalidOrderNumber                             = errors.New("invalid order number")
)
