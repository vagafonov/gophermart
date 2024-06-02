package convertor

import (
	"gophermart/internal/dto"
	"gophermart/internal/entity"
)

//nolint:exhaustive
func OrderStatusFromDtoToEntity(s dto.OrderAccrualStatus) entity.OrderStatus {
	switch s {
	case dto.OrderStatusRegistered:
		return entity.OrderStatusNew
	case dto.OrderStatusProcessing:
		return entity.OrderStatusProcessing
	case dto.OrderStatusProcessed:
		return entity.OrderStatusProcessed
	default:
		return entity.OrderStatusInvalid
	}
}
