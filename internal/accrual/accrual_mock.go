package accrual

import (
	"context"

	"gophermart/internal/dto"
	"gophermart/internal/entity"
)

type AccrualMock struct {
	getInfoResp *dto.OrderAccrual
}

func NewAccrualMock() *AccrualMock {
	return &AccrualMock{}
}

func (nam *AccrualMock) GetInfo(ctx context.Context, o entity.Order) (*dto.OrderAccrual, error) {
	return nam.getInfoResp, nil
}

func (nam *AccrualMock) SetInfo(o string, s dto.OrderAccrualStatus, a *float64) {
	nam.getInfoResp = &dto.OrderAccrual{
		Order:   o,
		Status:  s,
		Accrual: a,
	}
}
