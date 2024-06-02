package database

import (
	"context"
	"fmt"
	"gophermart/pkg/utils/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func Connect(ctx context.Context, lr *zap.SugaredLogger, dsn string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		lr.Error("failed connect to DB", zap.Error(err))
	}

	return dbpool
}

func Migrate(databaseURL string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", fs.Migrations("")),
		databaseURL)
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}

	return m.Up()
}
