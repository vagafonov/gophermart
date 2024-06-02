//nolint:funlen
package application

import (
	"context"
	"errors"
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

type UserOrderUploadTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestUserOrderUploadTestSuite(t *testing.T) {
	suite.Run(t, new(UserOrderUploadTestSuite))
}

func (s *UserOrderUploadTestSuite) SetupSuite() {
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

func (s *UserOrderUploadTestSuite) TearDownSuite() {
}

func (s *UserOrderUploadTestSuite) TestUserOrders() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("new order number has been accepted for processing", func() {
		n := goluhn.Generate(9)
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(n))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusAccepted, resp.StatusCode)
		<-s.app.needHandleNewOrder
	})

	s.Run("order number has already been uploaded by current user", func() {
		n := goluhn.Generate(9)
		userID := uuid.Must(uuid.NewUUID())
		_, err := s.cnt.GetOrderRepository().Add(ctx, n, userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 1, time.Now())
		s.Require().NoError(err)

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(n))
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

	s.Run("order number has already been uploaded by some user", func() {
		n := goluhn.Generate(9)
		someUserID := uuid.Must(uuid.NewUUID())
		userID := uuid.Must(uuid.NewUUID())
		_, err := s.cnt.GetOrderRepository().Add(ctx, n, someUserID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 1, time.Now())
		s.Require().NoError(err)

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(n))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("bad request", func() {
		req := ""
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(req))
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

	s.Run("user is not authenticated", func() {
		n := goluhn.Generate(9)

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(n))
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("incorrect order number format", func() {
		n := "12345678902"
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/orders", strings.NewReader(n))
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
