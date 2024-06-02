package contract

import (
	"context"

	"gophermart/internal/dto"
	"gophermart/internal/entity"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserIDFromJWT(tokenString string) (uuid.UUID, error)
	Create(ctx context.Context, l string, p string) (*entity.User, error)
	GetAuthToken(userID uuid.UUID) (string, error)
	GetByLoginAndPassword(ctx context.Context, p string, l string) (*entity.User, error)
}

type OrderService interface {
	AddNew(ctx context.Context, n string, uID uuid.UUID) (*entity.Order, error)
	UpdateStatusAndAccrual(ctx context.Context, n string, st entity.OrderStatus, a *float64) error
	GetByUser(ctx context.Context, uID uuid.UUID) ([]dto.Order, error)
	GetBalance(ctx context.Context, uID uuid.UUID) (*dto.Balance, error)
	WithdrawBalance(ctx context.Context, uID uuid.UUID, n string, a float64) (*entity.Order, error)
	GetListNewStatus(ctx context.Context, l int) ([]entity.Order, error)
	GetWithdrawals(ctx context.Context, uID uuid.UUID) ([]dto.Withdrawal, error)
}
