package service

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka/producer"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	pkgJwt "github.com/vogiaan1904/ticketbottle-order/pkg/jwt"
	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
)

type Config struct {
	PaymentTimeoutSeconds int32
}

type implService struct {
	l      logger.Logger
	cfg    Config
	repo   repo.Repository
	jwt    pkgJwt.Manager
	prod   producer.Producer
	invSvc inventory.InventoryServiceClient
	evSvc  event.EventServiceClient
	pmtSvc payment.PaymentServiceClient
}

func New(l logger.Logger, cfg Config, repo repo.Repository, invSvc inventory.InventoryServiceClient, evSvc event.EventServiceClient, pmtSvc payment.PaymentServiceClient, prod producer.Producer) Service {
	return &implService{
		l:      l,
		cfg:    cfg,
		repo:   repo,
		invSvc: invSvc,
		evSvc:  evSvc,
		pmtSvc: pmtSvc,
		prod:   prod,
	}
}
