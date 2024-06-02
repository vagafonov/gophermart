package request

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/errs"
	"io"
	"math"

	"github.com/ShiraazMoollatjie/goluhn"
)

type UserBalanceWithdraw struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func ValidateUserBalanceWithdraw(body io.ReadCloser) (*UserBalanceWithdraw, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	var req *UserBalanceWithdraw
	if err := json.Unmarshal(b, &req); err != nil {
		return nil, fmt.Errorf("cannot unmarshal request: %w", err)
	}

	if req.Order == "" {
		return nil, errs.NewValidationError("empty order number")
	}

	if req.Sum == 0 {
		return nil, errs.NewValidationError("empty sum")
	}

	req.Sum = 0 - math.Abs(req.Sum)

	err = goluhn.Validate(req.Order)
	if err != nil {
		return nil, errs.NewValidationError("invalid order number").Wrap(errs.ErrInvalidOrderNumber)
	}

	return req, nil
}
