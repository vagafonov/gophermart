package postgres

import (
	"context"
	"fmt"
	"gophermart/internal/contract"
	"gophermart/internal/entity"
	"gophermart/internal/errs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type UserPostgres struct {
	conn *pgxpool.Pool
}

func NewUserPostgres(conn *pgxpool.Pool) contract.UserRepository {
	return &UserPostgres{
		conn: conn,
	}
}

func (r *UserPostgres) Add(ctx context.Context, l string, p string) (*entity.User, error) {
	q := `INSERT INTO users (id, login, password) VALUES ($1, $2, $3)`
	id := uuid.New()
	_, err := r.conn.Exec(ctx, q, id, l, p)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user: %w", err)
	}

	return &entity.User{
		ID:       id,
		Login:    l,
		Password: p,
	}, nil
}

func (r *UserPostgres) GetByLogin(ctx context.Context, l string) (*entity.User, error) {
	var u entity.User
	q := `SELECT id, login, password FROM users WHERE login = $1`
	err := r.conn.QueryRow(ctx, q, l).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot select user by login: %w", err)
	}

	return &u, nil
}
