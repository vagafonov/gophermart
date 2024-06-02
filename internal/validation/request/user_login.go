package request

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/errs"
	"io"
)

type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func ValidateUserLogin(body io.ReadCloser) (*UserLogin, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	var req *UserLogin
	if err := json.Unmarshal(b, &req); err != nil {
		return nil, fmt.Errorf("cannot unmarshal request: %w", err)
	}

	if req.Login == "" || req.Password == "" {
		return req, errs.NewValidationError("empty login or password")
	}

	return req, nil
}
