//nolint:funlen
package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
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
	"gophermart/internal/logger"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"
)

type LoginTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

func (s *LoginTestSuite) SetupSuite() {
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

func (s *LoginTestSuite) TearDownSuite() {
}

func (s *LoginTestSuite) TestUserLogin() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successfully login", func() {
		l := fmt.Sprintf("gopher_%d", time.Now().UnixNano())
		_, err := s.cnt.GetUserService().Create(ctx, l, "123")
		s.Require().NoError(err)

		b := `{"login":"` + l + `","password":"123"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/login", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().NotEmpty(resp.Header.Get("Authorization"))
	})

	s.Run("user login not found", func() {
		b := `{"login":"not_found_login","password":"123"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/login", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("user password incorrect", func() {
		l := fmt.Sprintf("gopher_%d", time.Now().UnixNano())
		_, err := s.cnt.GetUserService().Create(ctx, l, "123")
		s.Require().NoError(err)
		b := `{"login":"` + l + `","password":"incorrect_password"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/login", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("bad request", func() {
		b := `{"someField": "someValue"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/login", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("internal server error", func() {
		b := ``
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/register", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusInternalServerError, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})
}
