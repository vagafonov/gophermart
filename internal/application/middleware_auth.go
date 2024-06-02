package application

import (
	"net/http"

	"github.com/google/uuid"
)

func (a *Application) middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("Authorization")
		if t == "" {
			w.WriteHeader(http.StatusUnauthorized)
			a.cnt.GetLogger().Info("authorization header is empty")

			return
		}

		userID, err := a.cnt.GetUserService().GetUserIDFromJWT(t)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			a.cnt.GetLogger().Errorf("failed to get user ID from JWT: %w", err)

			return
		}

		a.cnt.GetLogger().Infof("user ID: %v", userID)

		if userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			a.cnt.GetLogger().Info("user ID from JWT is empty")

			return
		}

		a.userID = userID
		next.ServeHTTP(w, r)
	})
}
