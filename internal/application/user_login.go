package application

import (
	"errors"
	"net/http"

	"gophermart/internal/errs"
	"gophermart/internal/validation/request"
)

func (a *Application) apiUserLogin(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("cal /api/user/login")
	req, err := request.ValidateUserLogin(r.Body)
	var target errs.ValidationError
	if err != nil {
		if errors.As(err, &target) {
			a.cnt.GetLogger().Errorf("validation failed %w", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		a.cnt.GetLogger().Errorf("validation failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	user, err := a.cnt.GetUserService().GetByLoginAndPassword(r.Context(), req.Password, req.Login)
	if err != nil {
		a.cnt.GetLogger().Errorf("get user failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if user == nil {
		a.cnt.GetLogger().Errorf("incorrect login or password %w", err)
		w.WriteHeader(http.StatusUnauthorized)

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
