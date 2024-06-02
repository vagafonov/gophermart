package application

import (
	"errors"
	"net/http"

	"gophermart/internal/errs"
	"gophermart/internal/validation/request"
)

func (a *Application) apiUserOrderUpload(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := request.ValidateUserOrders(r.Body)
	a.cnt.GetLogger().Info("call POST /api/user/orders order #" + orderNumber)

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
		a.cnt.GetLogger().Errorf("validation failed %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	_, err = a.cnt.GetOrderService().AddNew(r.Context(), orderNumber, a.userID)
	if err != nil {
		a.cnt.GetLogger().Errorf("add new order failed %w", err)

		switch {
		case errors.Is(err, errs.ErrOrderNumberHasAlreadyBeenUploadedBySomeUser):
			w.WriteHeader(http.StatusConflict)

			return
		case errors.Is(err, errs.ErrOrderNumberHasAlreadyBeenUploadedByCurrentUser):
			w.WriteHeader(http.StatusOK)

			return
		default:
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
	a.needHandleNewOrder <- true
}
