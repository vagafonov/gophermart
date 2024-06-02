package application

import (
	"encoding/json"
	"net/http"
)

func (a *Application) apiUserOrderGet(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("call GET /api/user/orders")
	w.Header().Set("Content-Type", "application/json")

	orders, err := a.cnt.GetOrderService().GetByUser(r.Context(), a.userID)
	if err != nil {
		a.cnt.GetLogger().Errorf("failed to get orders %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		a.cnt.GetLogger().Errorf("failed to encode orders %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
