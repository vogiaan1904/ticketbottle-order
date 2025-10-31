package activities

import (
	"context"
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
)

type OrderActivities struct {
	Repo repo.Repository
}

func NewOrderActivities(repo repo.Repository) *OrderActivities {
	return &OrderActivities{
		Repo: repo,
	}
}

func (a *OrderActivities) GetOrder(ctx context.Context, code string) (*models.Order, error) {
	o, err := a.Repo.GetOne(ctx, repo.GetOneOrderOption{
		FilterOrder: order.FilterOrder{
			Code: code,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &o, nil
}

func (a *OrderActivities) UpdateOrderStatus(ctx context.Context, ID string, status models.OrderStatus) error {
	_, err := a.Repo.Update(ctx, ID, repo.UpdateOrderOption{
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (a *OrderActivities) DeleteOrder(ctx context.Context, ID string) error {
	err := a.Repo.Delete(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (a *OrderActivities) DeleteOrderItems(ctx context.Context, ID string) error {
	err := a.Repo.DeleteItemByOrderID(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	return nil
}
