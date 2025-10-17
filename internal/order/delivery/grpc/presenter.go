package grpc

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/service"
	orderpb "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/order"
)

var GrpcOrderStatusValue = map[models.OrderStatus]orderpb.OrderStatus{
	models.OrderStatusPending:       orderpb.OrderStatus_ORDER_STATUS_PENDING,
	models.OrderStatusCompleted:     orderpb.OrderStatus_ORDER_STATUS_COMPLETED,
	models.OrderStatusCancelled:     orderpb.OrderStatus_ORDER_STATUS_CANCELED,
	models.OrderStatusPaymentFailed: orderpb.OrderStatus_ORDER_STATUS_FAILED,
}

func (s *GrpcService) newOrderItems(itms []models.OrderItem) []*orderpb.OrderItem {
	pbItems := make([]*orderpb.OrderItem, len(itms))
	for i, item := range itms {
		pbItems[i] = &orderpb.OrderItem{
			TicketClassId: item.TicketClassID,
			Quantity:      item.Quantity,
			PriceCents:    item.TotalAmount,
		}
	}

	return pbItems
}

func (s *GrpcService) newCreateResponses(out service.CreateOrderOutput) *orderpb.CreateOrderResponse {
	ord := &orderpb.Order{
		Id:               out.Order.ID.Hex(),
		Code:             out.Order.Code,
		UserId:           out.Order.UserID,
		EventId:          out.Order.EventID,
		UserFullname:     out.Order.UserFullName,
		UserEmail:        out.Order.Email,
		TotalAmountCents: out.Order.TotalAmount,
		Currency:         out.Order.Currency,
		PaymentMethod:    string(out.Order.PaymentMethod),
		Status:           GrpcOrderStatusValue[out.Order.Status],
		CreatedAt:        out.Order.CreatedAt.String(),
		UpdatedAt:        out.Order.UpdatedAt.String(),
		Items:            s.newOrderItems(out.OrderItems),
	}

	return &orderpb.CreateOrderResponse{
		Order:       ord,
		RedirectUrl: out.RedirectUrl,
	}
}
