package application

import (
	"errors"
	"net/http"

	"gophermart/internal/errs"
	"gophermart/internal/validation/request"
)

func (a *Application) apiUserRegister(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("cal /api/user/register")
	req, err := request.ValidateUserRegister(r.Body)
	var target errs.ValidationError
	if err != nil {
		if errors.As(err, &target) {
			a.cnt.GetLogger().Errorf("validation failed %w", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		a.cnt.GetLogger().Errorf("validation failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	user, err := a.cnt.GetUserService().Create(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrAlreadyExists) {
			a.cnt.GetLogger().Errorf("user already exists %w", err)
			w.WriteHeader(http.StatusConflict)

			return
		}

		a.cnt.GetLogger().Errorf("create user failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	t, err := a.cnt.GetUserService().GetAuthToken(user.ID)
	if err != nil {
		a.cnt.GetLogger().Errorf("create auth token failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	a.cnt.GetLogger().Infow("auth token create successfully", "token", t)
	w.Header().Set("Authorization", t)
	w.WriteHeader(http.StatusOK)
}
