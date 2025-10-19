package grpc

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	orderpb "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/order"
	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
	"github.com/vogiaan1904/ticketbottle-order/pkg/paginator"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcService struct {
	svc order.Service
	l   logger.Logger
	orderpb.UnimplementedOrderServiceServer
}

func NewGrpcService(svc order.Service, l logger.Logger) orderpb.OrderServiceServer {
	return &grpcService{
		svc: svc,
		l:   l,
	}
}

func (s *grpcService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if err := s.validateCreateOrderRequest(req); err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.Create.validateCreateOrderRequest: %v", err)
		return nil, err
	}

	in := order.CreateOrderInput{
		UserID:        req.UserId,
		EventID:       req.EventId,
		UserFullName:  req.UserFullname,
		Email:         req.UserEmail,
		Currency:      req.Currency,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
	}

	itms := make([]order.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		itms[i] = order.OrderItemInput{
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

func (s *grpcService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*emptypb.Empty, error) {
	if err := s.validateCancelOrderRequest(req); err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.Cancel.validateCancelOrderRequest: %v", err)
		return nil, err
	}

	err := s.svc.Cancel(ctx, req.Id)
	if err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.Cancel: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *grpcService) GetManyOrders(ctx context.Context, req *orderpb.GetManyOrdersRequest) (*orderpb.GetManyOrdersResponse, error) {
	if err := s.validateGetManyOrdersRequest(req); err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.GetManyOrders.validateGetManyOrdersRequest: %v", err)
		return nil, err
	}

	pagQ := paginator.PaginatorQuery{
		Page:  int(req.GetPage()),
		Limit: int64(req.GetPageSize()),
	}

	in := order.GetManyOrderInput{
		Pag: pagQ,
	}

	reqFil := req.GetFilter()
	if reqFil != nil {
		in.UserID = reqFil.GetUserId()
		in.EventID = reqFil.GetEventId()
		if reqFil.GetStatus() != 0 {
			stt := OrderStatus[reqFil.GetStatus()]
			in.Status = &stt
		}
	}

	out, err := s.svc.GetMany(ctx, in)
	if err != nil {
		s.l.Errorf(ctx, "internal.order.delivery.grpc.service.GetManyOrders: %v", err)
		return nil, err
	}

	return s.newGetManyOrdersResponse(out), nil
}
