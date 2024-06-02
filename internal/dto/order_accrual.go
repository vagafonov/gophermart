package dto

type OrderAccrualStatus string

//nolint:lll
const (
	OrderStatusRegistered OrderAccrualStatus = "REGISTERED" // заказ зарегистрирован, но вознаграждение не рассчитано
	OrderStatusProcessing OrderAccrualStatus = "PROCESSING" // расчёт начисления в процессе
	OrderStatusInvalid    OrderAccrualStatus = "INVALID"    // заказ не принят к расчёту, и вознаграждение не будет начислено
	OrderStatusProcessed  OrderAccrualStatus = "PROCESSED"  // расчёт начисления окончен
)

type OrderAccrual struct {
	Order   string             `json:"order"`
	Status  OrderAccrualStatus `json:"status"`
	Accrual *float64           `json:"accrual"`
}
