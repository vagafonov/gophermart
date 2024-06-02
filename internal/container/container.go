package container

import (
	"gophermart/internal/contract"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Container struct {
	l               *zap.SugaredLogger
	db              *pgxpool.Pool
	userRepository  contract.UserRepository
	userService     contract.UserService
	orderRepository contract.OrderRepository
	orderService    contract.OrderService
}

func NewContainer(
	l *zap.SugaredLogger,
	db *pgxpool.Pool,
	userRepository contract.UserRepository,
	userService contract.UserService,
	orderRepository contract.OrderRepository,
	orderService contract.OrderService,
) *Container {
	return &Container{
		l:               l,
		db:              db,
		userRepository:  userRepository,
		userService:     userService,
		orderRepository: orderRepository,
		orderService:    orderService,
	}
}

func (c *Container) GetLogger() *zap.SugaredLogger {
	return c.l
}

func (c *Container) GetDB() *pgxpool.Pool {
	return c.db
}

func (c *Container) GetUserRepository() contract.UserRepository {
	return c.userRepository
}

func (c *Container) GetUserService() contract.UserService {
	return c.userService
}

func (c *Container) GetOrderRepository() contract.OrderRepository {
	return c.orderRepository
}

func (c *Container) GetOrderService() contract.OrderService {
	return c.orderService
}
