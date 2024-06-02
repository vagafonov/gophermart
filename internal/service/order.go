package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gophermart/internal/contract"
	"gophermart/internal/dto"
	"gophermart/internal/entity"
	"gophermart/internal/errs"
)

type Order struct {
	orderRepo contract.OrderRepository
}

func NewOrder(orderRepo contract.OrderRepository) contract.OrderService {
	return &Order{orderRepo}
}

func (s *Order) AddNew(ctx context.Context, n string, uID uuid.UUID) (*entity.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, n)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if o != nil {
		if o.UserID == uID {
			return nil, errs.ErrOrderNumberHasAlreadyBeenUploadedByCurrentUser
		} else {
			return nil, errs.ErrOrderNumberHasAlreadyBeenUploadedBySomeUser
		}
	}

	return s.orderRepo.Add(ctx, n, uID, entity.OrderStatusNew, 0, 0, time.Now().UTC())
}

func (s *Order) GetByUser(ctx context.Context, uID uuid.UUID) ([]dto.Order, error) {
	orders, err := s.orderRepo.GetByUserIDAsc(ctx, uID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user ID: %w", err)
	}

	dtoOrders := make([]dto.Order, len(orders))
	for k, v := range orders {
		var accrual float64
		if v.Amount != 0 {
			accrual = v.Amount
		}
		dtoOrders[k] = dto.Order{
			ID:         v.ID,
			Status:     v.GetStringStatus(),
			Acrual:     &accrual,
			UploadedAt: v.CreatedAt,
		}
	}

	return dtoOrders, nil
}

func (s *Order) GetBalance(ctx context.Context, uID uuid.UUID) (*dto.Balance, error) {
	b, err := s.orderRepo.GetReplenishmentAndWithdrawalByUserID(ctx, uID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &dto.Balance{
		Current:   b.Replenishment,
		Withdrawn: math.Abs(b.Withdrawal),
	}, nil
}

func (s *Order) WithdrawBalance(ctx context.Context, uID uuid.UUID, n string, a float64) (*entity.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, n)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if o != nil {
		if o.UserID == uID {
			return nil, errs.ErrOrderNumberHasAlreadyBeenUploadedByCurrentUser
		} else {
			return nil, errs.ErrOrderNumberHasAlreadyBeenUploadedBySomeUser
		}
	}

	return s.orderRepo.AddWithdraw(ctx, n, uID, entity.OrderStatusProcessed, entity.OrderTypeWithdraw, a, time.Now().UTC())
}

func (s *Order) GetListNewStatus(ctx context.Context, l int) ([]entity.Order, error) {
	orders, err := s.orderRepo.GetList(ctx, entity.OrderStatusNew, l, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders with new status: %w", err)
	}

	return orders, nil
}

func (s *Order) UpdateStatusAndAccrual(ctx context.Context, n string, st entity.OrderStatus, a *float64) error {
	err := s.orderRepo.Update(ctx, n, st, a, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update order status with accural: %w", err)
	}

	return nil
}

func (s *Order) GetWithdrawals(ctx context.Context, uID uuid.UUID) ([]dto.Withdrawal, error) {
	ws, err := s.orderRepo.GetWithdrawalsByUserSortOld(ctx, uID, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals orders by user ID: %w", err)
	}

	dtoWithdrawals := make([]dto.Withdrawal, len(ws))
	for k, v := range ws {
		dtoWithdrawals[k] = dto.Withdrawal{
			Order:       v.ID,
			Sum:         math.Abs(v.Amount),
			ProcessedAt: v.CreatedAt,
		}
	}

	return dtoWithdrawals, nil
}
