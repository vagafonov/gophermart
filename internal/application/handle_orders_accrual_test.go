package application

import (
	"context"
	"errors"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"

	"gophermart/internal/accrual"
	"gophermart/internal/config"
	"gophermart/internal/container"
	"gophermart/internal/database"
	"gophermart/internal/dto"
	"gophermart/internal/entity"
	"gophermart/internal/logger"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"
)

type HandleOrdersAccrualTestSuite struct {
	suite.Suite
	cnt *container.Container
	app *Application
}

func TestHandleOrdersAccrualTestSuite(t *testing.T) {
	suite.Run(t, new(HandleOrdersAccrualTestSuite))
}

func (s *HandleOrdersAccrualTestSuite) SetupSuite() {
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

func (s *HandleOrdersAccrualTestSuite) TestHandleOrdersAccrualProducer() {
	ctx := context.Background()

	s.Run("p", func() {
		_, err := s.cnt.GetDB().Exec(ctx, "TRUNCATE orders")
		s.Require().NoError(err)
		userID := uuid.Must(uuid.NewUUID())
		orderExp, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 0, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		eg := new(errgroup.Group)
		eg.Go(func() error {
			s.Require().NoError(s.app.HandleOrdersAccrualProduccer(ctx))

			return nil
		})

		s.app.needHandleNewOrder <- true
		orderForConsume := <-s.app.chNewOrders
		s.Require().NoError(eg.Wait())
		s.Require().Equal(orderExp, orderForConsume)
	})
}

func (s *HandleOrdersAccrualTestSuite) TestHandleOrdersAccrualConsumer() {
	ctx := context.Background()

	s.Run("accrual return OrderStatusProcessing", func() {
		_, err := s.cnt.GetDB().Exec(ctx, "TRUNCATE orders")
		s.Require().NoError(err)
		userID := uuid.Must(uuid.NewUUID())
		orderDB, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		accrual := accrual.NewAccrualMock()
		accrual.SetInfo(orderDB.ID, dto.OrderStatusProcessing, nil)
		eg := new(errgroup.Group)
		eg.Go(func() error {
			s.Require().NoError(s.app.HandleOrdersAccrualConsumer(ctx, accrual))

			return nil
		})
		s.app.chNewOrders <- orderDB
		s.app.chNewOrders <- nil
		s.Require().NoError(eg.Wait())
		orderDB, err = s.cnt.GetOrderRepository().GetByID(ctx, orderDB.ID)
		s.Require().NoError(err)
		s.Require().Equal(entity.OrderStatusProcessing, orderDB.Status)
		s.Require().Equal(float64(0), orderDB.Amount) //nolint:testifylint
	})

	s.Run("accrual return OrderStatusProcessed", func() {
		_, err := s.cnt.GetDB().Exec(ctx, "TRUNCATE orders")
		s.Require().NoError(err)
		userID := uuid.Must(uuid.NewUUID())
		orderDB, err := s.cnt.GetOrderRepository().Add(ctx, goluhn.Generate(9), userID, entity.OrderStatusNew, entity.OrderTypeReplenishment, 1, time.Now().UTC().Truncate(time.Microsecond))
		s.Require().NoError(err)
		accrual := accrual.NewAccrualMock()
		amount := float64(1)
		accrual.SetInfo(orderDB.ID, dto.OrderStatusProcessed, &amount)
		eg := new(errgroup.Group)
		eg.Go(func() error {
			s.Require().NoError(s.app.HandleOrdersAccrualConsumer(ctx, accrual))

			return nil
		})
		s.app.chNewOrders <- orderDB
		s.app.chNewOrders <- nil
		s.Require().NoError(eg.Wait())
		orderDB, err = s.cnt.GetOrderRepository().GetByID(ctx, orderDB.ID)
		s.Require().NoError(err)
		s.Require().Equal(entity.OrderStatusProcessed, orderDB.Status)
		s.Require().Equal(float64(1), orderDB.Amount) //nolint:testifylint
	})
}
