package repository

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
)

type OrderItemRepository interface {
	CreateMany(ctx context.Context, ordID string, opts []CreateOrderItemOption) ([]models.OrderItem, error)
	ListByOrderID(ctx context.Context, ordID string) ([]models.OrderItem, error)
	DeleteByOrderID(ctx context.Context, ordID string) error
}
