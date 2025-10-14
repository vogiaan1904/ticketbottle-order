package repository

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, opt CreateOrderOption) (models.Order, error)
	GetByCode(ctx context.Context, code string) (models.Order, error)
	GetByID(ctx context.Context, ID string) (models.Order, error)
	Update(ctx context.Context, ID string, opt UpdateOrderOption) (models.Order, error)
	Delete(ctx context.Context, ID string) error
}
