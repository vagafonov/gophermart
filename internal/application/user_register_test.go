//nolint:funlen
package application

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strconv"
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

type UserRegisterTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestUserRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(UserRegisterTestSuite))
}

func (s *UserRegisterTestSuite) SetupSuite() {
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

func (s *UserRegisterTestSuite) TearDownSuite() {
}

func (s *UserRegisterTestSuite) TestUserRegister() {
	srv := httptest.NewServer(s.app.Routes())
	ctx := context.Background()

	s.Run("successfully registration", func() {
		b := `{"login":"gopher_` + strconv.FormatInt(time.Now().UnixNano(), 10) + `","password":"123"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/register", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().NotEmpty(resp.Header.Get("Authorization"))
	})

	// GenerateFromPassword не принимает пароли длиной более 72 байтов — это самый длинный пароль, с которым работает bcrypt.
	s.Run("too long password", func() {
		b := `{"login":"gopher_` + strconv.FormatInt(time.Now().UnixNano(), 10) + `","password":"` + strings.Repeat("*", 73) + `"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/register", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("login already exists", func() {
		login := "gopher_" + strconv.FormatInt(time.Now().UnixNano(), 10)
		_, err := s.cnt.GetUserRepository().Add(ctx, login, "123")
		s.Require().NoError(err)
		b := `{"login":"` + login + `","password":"123"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/register", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusConflict, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("bad request", func() {
		b := `{"someField": "someValue"}`
		r := httptest.NewRequest(http.MethodPost, srv.URL+"/api/user/register", strings.NewReader(b))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		s.Require().Empty(resp.Header.Get("Authorization"))
	})

	s.Run("bad request", func() {
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
