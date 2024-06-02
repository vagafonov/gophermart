package application

import (
	"encoding/json"
	"errors"
	"net/http"

	"gophermart/internal/errs"
	"gophermart/internal/validation/request"
)

func (a *Application) apiUserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("call POST /api/user/balance/withdraw")
	withdrawReq, err := request.ValidateUserBalanceWithdraw(r.Body)
	var target errs.ValidationError
	if err != nil {
		if errors.Is(err, errs.ErrInvalidOrderNumber) {
			a.cnt.GetLogger().Errorf("validation failed %w", err)
			w.WriteHeader(http.StatusUnprocessableEntity)

			return
		}

		if errors.As(err, &target) {
			a.cnt.GetLogger().Errorf("validation failed %w", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		a.cnt.GetLogger().Errorf("failed to check validation %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	orders, err := a.cnt.GetOrderService().WithdrawBalance(r.Context(), a.userID, withdrawReq.Order, withdrawReq.Sum)
	if err != nil {
		if errors.Is(err, errs.ErrInsufficientFundsOnBalance) {
			a.cnt.GetLogger().Errorf("failed check balance %w", err)
			w.WriteHeader(http.StatusPaymentRequired)
		}

		a.cnt.GetLogger().Errorf("failed to withdraw balance %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		a.cnt.GetLogger().Errorf("failed to encode orders %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.WriteHeader(http.StatusOK)
}
