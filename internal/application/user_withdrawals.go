package application

import (
	"encoding/json"
	"net/http"
)

func (a *Application) apiUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	a.cnt.GetLogger().Info("call GET /api/user/withdrawals")
	w.Header().Set("Content-Type", "application/json")

	withdrawals, err := a.cnt.GetOrderService().GetWithdrawals(r.Context(), a.userID)
	if err != nil {
		a.cnt.GetLogger().Errorf("failed to withdraw balance %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		a.cnt.GetLogger().Errorf("failed to encode orders %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.WriteHeader(http.StatusOK)
}
