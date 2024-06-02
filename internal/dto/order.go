package dto

import "time"

type Order struct {
	ID         string    `json:"number"`
	Status     string    `json:"status"`
	Acrual     *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"` //nolint:tagliatelle
}
