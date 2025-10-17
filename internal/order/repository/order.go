package repository

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
)

const (
	orderCollection = "orders"
)

func (r *implRepository) getOrderCollection() mongo.Collection {
	return r.db.Collection(orderCollection)
}

func (r *implRepository) Create(ctx context.Context, opt CreateOrderOption) (models.Order, error) {
	col := r.getOrderCollection()

	m := r.buildOrderModel(opt)
	if _, err := col.InsertOne(ctx, m); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Create: %v", err)
		return models.Order{}, err
	}

	return m, nil
}

func (r *implRepository) GetByCode(ctx context.Context, code string) (models.Order, error) {
	col := r.getOrderCollection()

	var m models.Order

	fil := r.buildGetByCodeQuery(ctx, code)
	if err := col.FindOne(ctx, fil).Decode(&m); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetByCode: %v", err)
		return models.Order{}, err
	}

	return m, nil
}

func (r *implRepository) GetByID(ctx context.Context, ID string) (models.Order, error) {
	col := r.getOrderCollection()

	var m models.Order

	fil, err := r.buildGetByIDQuery(ctx, ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetByID: %v", err)
		return models.Order{}, err
	}

	if err := col.FindOne(ctx, fil).Decode(&m); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetByID: %v", err)
		return models.Order{}, err
	}

	return m, nil
}

func (r *implRepository) Update(ctx context.Context, ID string, opt UpdateOrderOption) (models.Order, error) {
	col := r.getOrderCollection()

	m, upDoc := r.buildUpdateOrderModel(opt)

	fil, err := r.buildGetByIDQuery(ctx, ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Update: %v", err)
		return models.Order{}, err
	}

	if _, err := col.UpdateOne(ctx, fil, upDoc); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Update: %v", err)
		return models.Order{}, err
	}

	return m, nil
}

func (r *implRepository) Delete(ctx context.Context, ID string) error {
	col := r.getOrderCollection()

	fil, err := r.buildGetByIDQuery(ctx, ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Delete: %v", err)
		return err
	}

	if _, err := col.DeleteSoftOne(ctx, fil); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Delete: %v", err)
		return err
	}

	return nil
}
