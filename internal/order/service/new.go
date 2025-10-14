package service

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka/producer"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	pkgJwt "github.com/vogiaan1904/ticketbottle-order/pkg/jwt"
	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
)

type OrderService interface {
	Create(ctx context.Context, in CreateOrderInput) (CreateOrderOutput, error)

	Consumer
}

type Consumer interface {
	HandlePaymentStatus(ctx context.Context, in HandlePaymentStatusInput) error
}

type Config struct {
	PaymentTimeoutSeconds int32
}

type implOrderService struct {
	l      logger.Logger
	cfg    Config
	repo   repo.OrderRepository
	jwt    pkgJwt.Manager
	prod   producer.Producer
	itmSvc OrderItemService
	invSvc inventory.InventoryServiceClient
	evSvc  event.EventServiceClient
	pmtSvc payment.PaymentServiceClient
}

func NewOrderService(l logger.Logger, cfg Config, repo repo.OrderRepository, itmSvc OrderItemService, invSvc inventory.InventoryServiceClient, evSvc event.EventServiceClient, pmtSvc payment.PaymentServiceClient) OrderService {
	return &implOrderService{
		l:      l,
		cfg:    cfg,
		repo:   repo,
		itmSvc: itmSvc,
		invSvc: invSvc,
		evSvc:  evSvc,
		pmtSvc: pmtSvc,
	}
}
