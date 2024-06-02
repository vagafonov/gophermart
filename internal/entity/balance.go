package entity

import (
	"github.com/google/uuid"
)

type Balance struct {
	UserID        uuid.UUID
	Replenishment float64
	Withdrawal    float64
}
