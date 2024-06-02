package request

import (
	"fmt"
	"gophermart/internal/errs"
	"io"

	"github.com/ShiraazMoollatjie/goluhn"
)

func ValidateUserOrders(body io.ReadCloser) (string, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("cannot read request body: %w", err)
	}

	orderNumber := string(b)
	if orderNumber == "" {
		return "", errs.NewValidationError("empty order number")
	}
	err = goluhn.Validate(orderNumber)
	if err != nil {
		return "", errs.NewValidationError("invalid order number").Wrap(errs.ErrInvalidOrderNumber)
	}

	return orderNumber, nil
}
