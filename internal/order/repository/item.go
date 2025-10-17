package repository

import (
	"context"
	"sync"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	orderItemCollection = "order_items"
)

func (r *implRepository) getOrderItemCollection() mongo.Collection {
	return r.db.Collection(orderItemCollection)
}

func (r *implRepository) CreateManyItems(ctx context.Context, ordID string, opts []CreateOrderItemOption) ([]models.OrderItem, error) {
	col := r.getOrderItemCollection()

	itmsCh := make(chan models.OrderItem, len(opts))
	wg := sync.WaitGroup{}

	for i, opt := range opts {
		wg.Add(1)
		go func(i int, opt CreateOrderItemOption) {
			defer wg.Done()
			m := r.buildOrderItemModel(ordID, opt)
			itmsCh <- m
		}(i, opt)
	}

	wg.Wait()
	close(itmsCh)

	itms := make([]models.OrderItem, 0, len(opts))
	for itm := range itmsCh {
		itms = append(itms, itm)
	}

	docs := make([]interface{}, len(itms))
	for i, itm := range itms {
		docs[i] = itm
	}

	if _, err := col.InsertMany(ctx, docs); err != nil {
		r.l.Errorf(ctx, "order.reporitory.OrderItemRepository.CreateMany: %v", err)
		return nil, err
	}

	return itms, nil
}

func (r *implRepository) ListItemByOrderID(ctx context.Context, ordID string) ([]models.OrderItem, error) {
	col := r.getOrderItemCollection()

	oID, err := primitive.ObjectIDFromHex(ordID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderItemRepository.ListByOrderID: %v", err)
		return nil, err
	}

	cur, err := col.Find(ctx, bson.M{"order_id": oID})
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderItemRepository.ListByOrderID: %v", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var itms []models.OrderItem
	if err := cur.All(ctx, &itms); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderItemRepository.ListByOrderID: %v", err)
		return nil, err
	}

	return itms, nil
}

func (r *implRepository) DeleteItemByOrderID(ctx context.Context, ordID string) error {
	col := r.getOrderItemCollection()

	oID, err := primitive.ObjectIDFromHex(ordID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderItemRepository.DeleteByOrderID: %v", err)
		return err
	}

	_, err = col.DeleteMany(ctx, bson.M{"order_id": oID})
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderItemRepository.DeleteByOrderID: %v", err)
		return err
	}

	return nil
}
