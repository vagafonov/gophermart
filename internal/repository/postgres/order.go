package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"time"

	"gophermart/internal/contract"
	"gophermart/internal/entity"
	"gophermart/internal/errs"
)

type OrderPostgres struct {
	conn *pgxpool.Pool
}

func NewOrderPostgres(conn *pgxpool.Pool) contract.OrderRepository {
	return &OrderPostgres{
		conn: conn,
	}
}

func (r *OrderPostgres) Add(ctx context.Context, n string, uID uuid.UUID, s entity.OrderStatus, t entity.OrderType, a float64, cAt time.Time) (*entity.Order, error) {
	q := `INSERT INTO orders (id, user_id, status, type, amount, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, status, created_at, updated_at`
	_, err := r.conn.Exec(ctx, q, n, uID, s, t, a, cAt, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot insert order: %w", err)
	}

	return &entity.Order{
		ID:        n,
		UserID:    uID,
		Status:    s,
		Type:      t,
		Amount:    a,
		CreatedAt: cAt,
		UpdatedAt: nil,
	}, nil
}

func (r *OrderPostgres) GetByID(ctx context.Context, n string) (*entity.Order, error) {
	var o entity.Order
	q := `SELECT * FROM orders WHERE id = $1`
	err := r.conn.QueryRow(ctx, q, n).Scan(&o.ID, &o.UserID, &o.Status, &o.Type, &o.Amount, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot select order by ID: %w", err)
	}

	return &o, nil
}

func (r *OrderPostgres) GetByUserIDAsc(ctx context.Context, uID uuid.UUID) ([]entity.Order, error) {
	q := `SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at ASC LIMIT 200`
	rows, err := r.conn.Query(ctx, q, uID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot select order by ID: %w", err)
	}

	var o []entity.Order

	for rows.Next() {
		e := entity.Order{
			ID:        "",
			UserID:    uuid.UUID{},
			Status:    0,
			CreatedAt: time.Time{},
			UpdatedAt: nil,
		}
		err = rows.Scan(&e.ID, &e.UserID, &e.Status, &e.Type, &e.Amount, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("database row scan failed: %w", err)
		}
		o = append(o, e)
	}
	rows.Close()

	return o, nil
}

func (r *OrderPostgres) GetReplenishmentAndWithdrawalByUserID(ctx context.Context, uID uuid.UUID) (*entity.Balance, error) {
	var b entity.Balance
	q := `SELECT
	SUM(amount) AS total_sum,
	SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END) AS negative_sum
	FROM orders
	WHERE
    	user_id = $1;`

	err := r.conn.QueryRow(ctx, q, uID).Scan(&b.Replenishment, &b.Withdrawal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot select user balance: %w", err)
	}

	return &b, nil
}

func (r *OrderPostgres) GetList(ctx context.Context, s entity.OrderStatus, limit int, offset int) ([]entity.Order, error) {
	q := `SELECT * FROM orders WHERE status = $1 LIMIT $2 OFFSET $3`
	rows, err := r.conn.Query(ctx, q, s, limit, offset)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot get list: %w", err)
	}
	var o []entity.Order
	var oItem entity.Order
	for rows.Next() {
		err = rows.Scan(&oItem.ID, &oItem.UserID, &oItem.Status, &oItem.Type, &oItem.Amount, &oItem.CreatedAt, &oItem.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("database row scan failed: %w", err)
		}
		o = append(o, oItem)
	}
	rows.Close()

	return o, nil
}

func (r *OrderPostgres) Update(ctx context.Context, n string, s entity.OrderStatus, a *float64, uAt time.Time) error {
	q := `UPDATE orders SET status = $1, amount = $2, updated_at = $3 WHERE id = $4`
	var amount float64
	if a == nil {
		amount = 0
	} else {
		amount = *a
	}
	_, err := r.conn.Exec(ctx, q, s, amount, uAt, n)
	if err != nil {
		return fmt.Errorf("cannot update order: %w", err)
	}

	return nil
}

func (r *OrderPostgres) AddWithdraw(ctx context.Context, n string, uID uuid.UUID, s entity.OrderStatus, t entity.OrderType, a float64, cAt time.Time) (*entity.Order, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, hashtext($2));`, t, uID); err != nil {
		return nil, err
	}

	var isEnough bool
	q := `SELECT coalesce(SUM(amount) > $1, false) FROM orders WHERE user_id = $2;`
	err = tx.QueryRow(ctx, q, math.Abs(a), uID).Scan(&isEnough)
	if err != nil {
		return nil, fmt.Errorf("cannot check balance: %w", err)
	}
	if !isEnough {
		return nil, errs.ErrInsufficientFundsOnBalance
	}

	q = `INSERT INTO orders (id, user_id, status, type, amount, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	_, err = r.conn.Exec(ctx, q, n, uID, s, t, a, cAt, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot insert withdraw order: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err // TODO
	}

	return &entity.Order{
		ID:        n,
		UserID:    uID,
		Status:    s,
		Type:      t,
		Amount:    a,
		CreatedAt: cAt,
		UpdatedAt: nil,
	}, nil
}

func (r *OrderPostgres) GetWithdrawalsByUserSortOld(ctx context.Context, uID uuid.UUID, s int) ([]entity.Order, error) {
	sqlSort := "ASC"
	if s < 0 {
		sqlSort = "DESC"
	}
	q := fmt.Sprintf(`SELECT * FROM orders WHERE user_id = $1 AND type = $2 ORDER BY created_at %s LIMIT 200`, sqlSort)
	rows, err := r.conn.Query(ctx, q, uID, entity.OrderTypeWithdraw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}

		return nil, fmt.Errorf("cannot select withdrawals orders by userID: %w", err)
	}

	var o []entity.Order
	for rows.Next() {
		e := entity.Order{}
		err = rows.Scan(&e.ID, &e.UserID, &e.Status, &e.Type, &e.Amount, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("database row scan failed: %w", err)
		}
		o = append(o, e)
	}
	rows.Close()

	return o, nil
}
