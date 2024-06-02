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

type UserBalanceGetTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestUserBalanceGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserBalanceGetTestSuite))
}

func (s *UserBalanceGetTestSuite) SetupSuite() {
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

func (s *UserBalanceGetTestSuite) TestUserBalance() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successful request processing", func() {
		userID := uuid.Must(uuid.NewUUID())
		_, err := s.cnt.GetOrderRepository().Add(
			ctx,
			goluhn.Generate(9),
			userID,
			entity.OrderStatusProcessed,
			entity.OrderTypeReplenishment,
			2,
			time.Now().UTC().Truncate(time.Microsecond),
		)
		s.Require().NoError(err)
		_, err = s.cnt.GetOrderRepository().Add(
			ctx,
			goluhn.Generate(9),
			userID,
			entity.OrderStatusProcessed,
			entity.OrderTypeWithdraw,
			-1,
			time.Now().UTC().Truncate(time.Microsecond),
		)
		s.Require().NoError(err)
		_, err = s.cnt.GetOrderRepository().Add(
			ctx,
			goluhn.Generate(9),
			userID,
			entity.OrderStatusProcessed,
			entity.OrderTypeWithdraw,
			-1,
			time.Now().UTC().Truncate(time.Microsecond),
		)
		s.Require().NoError(err)

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/balance", strings.NewReader(""))
		t, err := s.cnt.GetUserService().GetAuthToken(userID)
		s.Require().NoError(err)

		r.RequestURI = ""
		r.Header["Authorization"] = []string{t}
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)
		var b dto.Balance
		err = json.Unmarshal(body, &b)
		s.Require().NoError(err)

		s.Require().Equal(dto.Balance{
			Current:   0,
			Withdrawn: 2,
		}, b)
	})

	s.Run("user is not authenticated", func() {
		n := goluhn.Generate(9)

		r := httptest.NewRequest(http.MethodGet, srv.URL+"/api/user/balance", strings.NewReader(n))
		r.RequestURI = ""
		r.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	})
}
