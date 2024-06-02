package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"

	"gophermart/internal/container"
	"gophermart/internal/contract"
	"gophermart/internal/convertor"
	"gophermart/internal/entity"
)

const limitGetListNewStatus = 100

type Application struct {
	cnt *container.Container
	container.Container
	userID                    uuid.UUID
	needHandleNewOrder        chan bool
	stopOrdersAccrualProducer chan struct{}
	chNewOrders               chan *entity.Order
}

func NewApplication(cnt *container.Container) *Application {
	return &Application{
		cnt:                       cnt,
		needHandleNewOrder:        make(chan bool, 1),
		stopOrdersAccrualProducer: make(chan struct{}),
		chNewOrders:               make(chan *entity.Order),
	}
}

func (a *Application) HandleOrdersAccrualProduccer(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-a.needHandleNewOrder:
				a.cnt.GetLogger().Info("new order created")
				orders, _ := a.cnt.GetOrderService().GetListNewStatus(ctx, limitGetListNewStatus)
				for _, v := range orders {
					a.chNewOrders <- &v //nolint:gosec
				}
			case <-a.stopOrdersAccrualProducer:
				a.cnt.GetLogger().Info("geeting signal for stop produccer")
				close(a.chNewOrders)

				return
			}
		}
	}()
	a.cnt.GetLogger().Infof("producer stopped")

	return nil
}

func (a *Application) HandleOrdersAccrualConsumer(ctx context.Context, ac contract.Accrual) error {
	for o := range a.chNewOrders {
		if o == nil {
			break
		}

		a.cnt.GetLogger().Infof("consumer handling order №:%s", o.ID)

		ordDTO, err := ac.GetInfo(ctx, *o)
		if err != nil {
			a.cnt.GetLogger().Errorf("cannot call /api/orders/ to accural: %w", err)
		}

		if ordDTO != nil {
			err = a.cnt.GetOrderService().UpdateStatusAndAccrual(
				ctx,
				ordDTO.Order,
				convertor.OrderStatusFromDtoToEntity(ordDTO.Status),
				ordDTO.Accrual,
			)
			if err != nil {
				return err
			}
		}
	}
	a.cnt.GetLogger().Infof("consumer stopped")

	return nil
}

func (a *Application) Serve() error {
	a.cnt.GetLogger().Infof("server started and listen %s", viper.GetString("run_address"))
	err := http.ListenAndServe(viper.GetString("run_address"), a.Routes()) //nolint:gosec
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) Routes() *mux.Router {
	r := mux.NewRouter()

	s := r.PathPrefix("/api/user/").Subrouter()
	s.HandleFunc("/register", a.apiUserRegister).Methods("POST")
	s.HandleFunc("/login", a.apiUserLogin).Methods("POST")

	o := r.PathPrefix("/api/user/").Subrouter()
	o.Use(a.middlewareAuth)
	o.HandleFunc("/orders", a.apiUserOrderUpload).Methods("POST")
	o.HandleFunc("/orders", a.apiUserOrderGet).Methods("GET")
	o.HandleFunc("/balance", a.apiUserBalanceGet).Methods("GET")
	o.HandleFunc("/balance/withdraw", a.apiUserBalanceWithdraw).Methods("POST")
	o.HandleFunc("/withdrawals", a.apiUserWithdrawals).Methods("GET")

	return r
}
