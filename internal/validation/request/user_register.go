package request

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/errs"
	"io"
)

type UserRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

const maxPasswordLength = 72

func ValidateUserRegister(body io.ReadCloser) (*UserRegister, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	var req *UserRegister
	if err := json.Unmarshal(b, &req); err != nil {
		return nil, fmt.Errorf("cannot unmarshal request: %w", err)
	}

	if req.Login == "" || req.Password == "" {
		return req, errs.NewValidationError("empty login or password")
	}

	if len(req.Password) > maxPasswordLength {
		return req, errs.NewValidationError(fmt.Sprintf("password must be less than :%d", maxPasswordLength))
	}

	return req, nil
}
