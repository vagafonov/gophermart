package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io"
	"net/http"

	"gophermart/internal/contract"
	"gophermart/internal/dto"
	"gophermart/internal/entity"
)

type Accrual struct {
	l *zap.SugaredLogger
}

func NewAccrual(l *zap.SugaredLogger) contract.Accrual {
	return &Accrual{
		l: l,
	}
}

func (a *Accrual) GetInfo(ctx context.Context, o entity.Order) (*dto.OrderAccrual, error) {
	addr := viper.GetString("accrual_system_address")
	a.l.Infof("call accural %s", addr+"/api/orders/"+o.ID)
	resp, err := http.Get(addr + "/api/orders/" + o.ID) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var o *dto.OrderAccrual
		if err := json.Unmarshal(b, &o); err != nil {
			return nil, fmt.Errorf("cannot unmarshal order: %w", err)
		}

		return o, nil
	}

	return nil, nil //nolint:nilnil
}
