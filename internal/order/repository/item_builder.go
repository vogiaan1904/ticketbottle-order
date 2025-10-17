package repository

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
)

func (r *implRepository) buildOrderItemModel(ordID string, opt CreateOrderItemOption) models.OrderItem {
	now := r.clock()
	m := models.OrderItem{
		ID:              r.db.NewObjectID(),
		OrderID:         mongo.ObjectIDFromHexOrNil(ordID),
		TicketClassID:   opt.TicketClassID,
		TicketClassName: opt.TicketClassName,
		PriceAtPurchase: opt.PriceAtPurchase,
		Quantity:        opt.Quantity,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return m
}
