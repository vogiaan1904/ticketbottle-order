package service

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
)

type OrderItemService interface {
	CreateMany(ctx context.Context, ordID string, items []CreateOrderItemInput) ([]models.OrderItem, error)
	DeleteByOrderID(ctx context.Context, ordID string) error
}

type implOrderItemService struct {
	repo repo.OrderItemRepository
}

func NewOrderItemService(repo repo.OrderItemRepository) OrderItemService {
	return &implOrderItemService{
		repo: repo,
	}
}

func (s *implOrderItemService) CreateMany(ctx context.Context, ordID string, items []CreateOrderItemInput) ([]models.OrderItem, error) {
	opts := make([]repo.CreateOrderItemOption, len(items))
	for i, item := range items {
		opts[i] = repo.CreateOrderItemOption{
			OrderID:         ordID,
			TicketClassID:   item.TicketClassID,
			TicketClassName: item.TicketClassName,
			PriceAtPurchase: item.PriceAtPurchase,
			Quantity:        item.Quantity,
			TotalAmount:     item.PriceAtPurchase * int64(item.Quantity),
		}
	}

	itms, err := s.repo.CreateMany(ctx, ordID, opts)
	if err != nil {
		return nil, err
	}

	return itms, nil
}

func (s *implOrderItemService) DeleteByOrderID(ctx context.Context, ordID string) error {
	return s.repo.DeleteByOrderID(ctx, ordID)
}
