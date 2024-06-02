//nolint:funlen
package application

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gophermart/internal/config"
	"gophermart/internal/container"
	"gophermart/internal/database"
	"gophermart/internal/dto"
	"gophermart/internal/entity"
	"gophermart/internal/logger"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"
)

type UserOrderGetTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestUserOrderGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserOrderGetTestSuite))
}

func (s *UserOrderGetTestSuite) SetupSuite() {
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

func (s *UserOrderGetTestSuite) TearDownSuite() {
}

func (s *UserOrderGetTestSuite) TestUserOrders() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successful request processing", func() {
		userID := uuid.Must(uuid.NewUUID())
		o1, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 123.45, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		o2, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusInvalid, entity.OrderTypeReplenishment, 1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		o3, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusProcessed, entity.OrderTypeReplenishment, 1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		o4, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusProcessing, entity.OrderTypeReplenishment, 1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/orders", strings.NewReader(""))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var orders []dto.Order
		err = json.Unmarshal(body, &orders)
		s.Require().NoError(err)
		o1Amount := o1.Amount
		oneAmount := float64(1)
		exp := []dto.Order{
			{
				ID:         o1.ID,
				Status:     o1.GetStringStatus(),
				Acrual:     &o1Amount,
				UploadedAt: o1.CreatedAt,
			},
			{
				ID:         o2.ID,
				Status:     o2.GetStringStatus(),
				Acrual:     &oneAmount,
				UploadedAt: o2.CreatedAt,
			},
			{
				ID:         o3.ID,
				Status:     o3.GetStringStatus(),
				Acrual:     &oneAmount,
				UploadedAt: o3.CreatedAt,
			},
			{
				ID:         o4.ID,
				Status:     o4.GetStringStatus(),
				Acrual:     &oneAmount,
				UploadedAt: o4.CreatedAt,
			},
		}

		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Len(orders, 4)
		s.Require().Equal(exp, orders)
	})

	s.Run("empty orders", func() {
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/orders", strings.NewReader(""))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().NoError(err)
		s.Require().Equal(http.StatusNoContent, resp.StatusCode)
	})

	s.Run("user is not authenticated", func() {
		n := goluhn.Generate(9)

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/orders", strings.NewReader(n))
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	})
}
