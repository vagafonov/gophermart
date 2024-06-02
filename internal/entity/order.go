package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus int16

const (
	OrderStatusNew        OrderStatus = iota + 1 // заказ загружен в систему, но не попал в обработку;
	OrderStatusProcessing                        // вознаграждение за заказ рассчитывается
	OrderStatusInvalid                           // система расчёта вознаграждений отказала в расчёте
	OrderStatusProcessed                         // данные по заказу проверены и информация о расчёте успешно получена
)

type OrderType int16

const (
	OrderTypeWithdraw      = -1 // Списание с баланса
	OrderTypeReplenishment = 1  // Пополнение баланса
)

//nolint:musttag
type Order struct {
	ID        string
	UserID    uuid.UUID
	Status    OrderStatus
	Type      OrderType
	Amount    float64
	CreatedAt time.Time
	UpdatedAt *time.Time
}

//nolint:exhaustive
func (o *Order) GetStringStatus() string {
	switch o.Status {
	case OrderStatusNew:
		return "NEW"
	case OrderStatusProcessing:
		return "PROCESSING"
	case OrderStatusProcessed:
		return "PROCESSED"
	default:
		return "INVALID"
	}
}
