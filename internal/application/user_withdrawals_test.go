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
	"math"
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

type UserWithdrawalsTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(UserWithdrawalsTestSuite))
}

func (s *UserWithdrawalsTestSuite) SetupSuite() {
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

func (s *UserWithdrawalsTestSuite) TearDownSuite() {
}

func (s *UserWithdrawalsTestSuite) TestUserOrders() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successful request processing", func() {
		userID := uuid.Must(uuid.NewUUID())
		o1, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeWithdraw, 10, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		o2, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusProcessed, entity.OrderTypeWithdraw, -1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		o3, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusProcessed, entity.OrderTypeWithdraw, -2, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/withdrawals", strings.NewReader(""))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)
		r.RequestURI = ""
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var ws []dto.Withdrawal
		err = json.Unmarshal(body, &ws)
		s.Require().NoError(err)

		exp := []dto.Withdrawal{
			{
				Order:       o3.ID,
				Sum:         math.Abs(o3.Amount),
				ProcessedAt: o3.CreatedAt,
			},
			{
				Order:       o2.ID,
				Sum:         math.Abs(o2.Amount),
				ProcessedAt: o2.CreatedAt,
			},
			{
				Order:       o1.ID,
				Sum:         math.Abs(o1.Amount),
				ProcessedAt: o1.CreatedAt,
			},
		}

		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Len(ws, 3)
		s.Require().Equal(exp, ws)
	})

	s.Run("empty withdrawals", func() {
		userID := uuid.Must(uuid.NewUUID())

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/withdrawals", strings.NewReader(""))
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
		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/withdrawals", strings.NewReader(""))
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	})
}
