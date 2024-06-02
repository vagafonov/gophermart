//nolint:funlen
package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gophermart/internal/config"
	"gophermart/internal/container"
	"gophermart/internal/database"
	"gophermart/internal/entity"
	"gophermart/internal/logger"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"
)

type UserOrderWithdrawTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestUserOrderWithdrawTestSuite(t *testing.T) {
	suite.Run(t, new(UserOrderWithdrawTestSuite))
}

func (s *UserOrderWithdrawTestSuite) SetupSuite() {
	ctx := context.Background()
	lr := logger.NewLogger()
	config.Init(lr, "test")
	config.DebugConfig(lr)
	pool := database.Connect(ctx, lr, viper.GetString("database_uri"))
	err := database.Migrate(viper.GetString("database_uri"))
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		lr.Errorf("failed to run migrations %w", err)
	}
	userReository := postgres.NewUserPostgres(pool)
	orderReository := postgres.NewOrderPostgres(pool)
	s.cnt = container.NewContainer(
		lr,
		pool,
		userReository,
		service.NewUser(userReository),
		orderReository,
		service.NewOrder(orderReository),
	)
	s.app = NewApplication(s.cnt)
}

func (s *UserOrderWithdrawTestSuite) TearDownSuite() {
}

func (s *UserOrderWithdrawTestSuite) TestOrderWithdraw() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successfully withdraw", func() {
		userID := uuid.Must(uuid.NewUUID())
		_, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 124, time.Now())
		s.Require().NoError(err)

		b := fmt.Sprintf(`{"order": "%s", "sum": %.2f}`, goluhn.Generate(9), -123.45)
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/balance/withdraw", strings.NewReader(b))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("user is not authenticated", func() {
		b := fmt.Sprintf(`{"order": "%s", "sum": %.2f}`, goluhn.Generate(9), 123.45)
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/balance/withdraw", strings.NewReader(b))
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("insufficient funds on balance", func() {
		n := goluhn.Generate(9)
		userID := uuid.Must(uuid.NewUUID())
		b := fmt.Sprintf(`{"order": "%s", "sum": %.2f}`, n, -123.45)
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/balance/withdraw", strings.NewReader(b))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusPaymentRequired, resp.StatusCode)
	})

	s.Run("bad request", func() {
		req := `{}`
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/balance/withdraw", strings.NewReader(req))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("incorrect order number format", func() {
		req := fmt.Sprintf(`{"order": "12345678902", "sum": %.2f}`, -123.45)
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/balance/withdraw", strings.NewReader(req))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	})
}
