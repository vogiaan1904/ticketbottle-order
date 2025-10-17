package grpc

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/service"
	orderpb "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/order"
	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
)

type GrpcService struct {
	svc service.Service
	l   logger.Logger
	orderpb.UnimplementedOrderServiceServer
}

func NewGrpcService(svc service.Service, l logger.Logger) *GrpcService {
	return &GrpcService{
		svc: svc,
		l:   l,
	}
}

func (s *GrpcService) Create(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	in := service.CreateOrderInput{
		UserID:        req.UserId,
		EventID:       req.EventId,
		UserFullName:  req.UserFullname,
		Email:         req.UserEmail,
		Currency:      req.Currency,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
	}

	itms := make([]service.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		itms[i] = service.OrderItemInput{
			TicketClassID: item.TicketClassId,
			Quantity:      item.Quantity,
		}
	}
	in.Items = itms

	out, err := s.svc.Create(ctx, in)
	if err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.Create: %v", err)
		return nil, err
	}

	return s.newCreateResponses(out), nil
}
