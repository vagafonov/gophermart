package contract

import (
	"context"

	"gophermart/internal/dto"
	"gophermart/internal/entity"
)

type Accrual interface {
	GetInfo(ctx context.Context, order entity.Order) (*dto.OrderAccrual, error)
}
