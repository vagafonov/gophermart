package application

import (
	"encoding/json"
	"net/http"
)

func (a *Application) apiUserBalanceGet(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("call GET /api/user/balance")
	w.Header().Set("Content-Type", "application/json")

	orders, err := a.cnt.GetOrderService().GetBalance(r.Context(), a.userID)
	if err != nil {
		a.cnt.GetLogger().Errorf("failed to get balance %w", err)
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
