package main

import (
	"context"
	"gophermart/internal/accrual"
	"gophermart/internal/application"
	"gophermart/internal/config"
	"gophermart/internal/container"
	"gophermart/internal/database"
	"gophermart/internal/logger"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	lr := logger.NewLogger()
	config.Init(lr, "config")
	config.ParseFlags(lr)
	config.DebugConfig(lr)
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	pool := database.Connect(ctx, lr, viper.GetString("database_uri"))
	defer pool.Close()

	err := database.Migrate(viper.GetString("database_uri"))
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		lr.Error("failed to run migrations", err)
	}

	userReository := postgres.NewUserPostgres(pool)
	orderReository := postgres.NewOrderPostgres(pool)
	cnt := container.NewContainer(
		lr,
		pool,
		userReository,
		service.NewUser(userReository),
		orderReository,
		service.NewOrder(orderReository),
	)

	app := application.NewApplication(cnt)
	go app.HandleOrdersAccrualProduccer(ctx) //nolint:errcheck
	ac := accrual.NewAccrual(cnt.GetLogger())
	go app.HandleOrdersAccrualConsumer(ctx, ac) //nolint:errcheck
	if err := app.Serve(); err != nil {
		lr.Error("failed to run application", zap.Error(err))
	}
}
