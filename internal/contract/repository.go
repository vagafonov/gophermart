package contract

import (
	"context"
	"github.com/google/uuid"
	"time"

	"gophermart/internal/entity"
)

type UserRepository interface {
	Add(ctx context.Context, l string, p string) (*entity.User, error)
	GetByLogin(ctx context.Context, l string) (*entity.User, error)
}

type OrderRepository interface {
	Add(ctx context.Context, n string, uID uuid.UUID, s entity.OrderStatus, t entity.OrderType, a float64, cAt time.Time) (*entity.Order, error)
	Update(ctx context.Context, n string, s entity.OrderStatus, a *float64, uAt time.Time) error
	GetByID(ctx context.Context, n string) (*entity.Order, error)
	GetByUserIDAsc(ctx context.Context, uID uuid.UUID) ([]entity.Order, error)
	GetReplenishmentAndWithdrawalByUserID(ctx context.Context, uID uuid.UUID) (*entity.Balance, error)
	GetList(ctx context.Context, s entity.OrderStatus, limit int, offset int) ([]entity.Order, error)
	AddWithdraw(ctx context.Context, n string, uID uuid.UUID, s entity.OrderStatus, t entity.OrderType, a float64, cAt time.Time) (*entity.Order, error)
	GetWithdrawalsByUserSortOld(ctx context.Context, uID uuid.UUID, s int) ([]entity.Order, error)
}
