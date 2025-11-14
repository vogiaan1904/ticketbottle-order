package activities

import (
	"context"

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

func (a *OrderActivities) CreateOrder(ctx context.Context, opt repo.CreateOrderOption) (*models.Order, error) {
	o, err := a.Repo.Create(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (a *OrderActivities) CreateOrderItems(ctx context.Context, orderID string, items []repo.CreateOrderItemOption) ([]models.OrderItem, error) {
	itms, err := a.Repo.CreateManyItems(ctx, orderID, items)
	if err != nil {
		return nil, err
	}

	return itms, nil
}

func (a *OrderActivities) GetOrder(ctx context.Context, code string) (*models.Order, error) {
	o, err := a.Repo.GetOne(ctx, repo.GetOneOrderOption{
		FilterOrder: order.FilterOrder{
			Code: code,
		},
	})
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (a *OrderActivities) UpdateOrderStatus(ctx context.Context, ID string, status models.OrderStatus) error {
	_, err := a.Repo.Update(ctx, ID, repo.UpdateOrderOption{
		Status: status,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *OrderActivities) DeleteOrder(ctx context.Context, ID string) error {
	err := a.Repo.Delete(ctx, ID)
	if err != nil {
		return err
	}

	return nil
}

func (a *OrderActivities) DeleteOrderItems(ctx context.Context, ID string) error {
	err := a.Repo.DeleteItemByOrderID(ctx, ID)
	if err != nil {
		return err
	}

	return nil
}
