package repository

import (
	"context"
	"sync"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
	"github.com/vogiaan1904/ticketbottle-order/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	orderCollection = "orders"
)

func (r *implRepository) getOrderCollection() mongo.Collection {
	return r.db.Collection(orderCollection)
}

func (r *implRepository) Create(ctx context.Context, opt CreateOrderOption) (models.Order, error) {
	col := r.getOrderCollection()

	o := r.buildOrderModel(opt)
	if _, err := col.InsertOne(ctx, o); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Create: %v", err)
		return models.Order{}, err
	}

	return o, nil
}

func (r *implRepository) GetOne(ctx context.Context, opt GetOneOrderOption) (models.Order, error) {
	col := r.getOrderCollection()

	q := r.buildFilterQuery(ctx, opt.FilterOrder)

	var o models.Order
	if err := col.FindOne(ctx, q).Decode(&o); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetOne: %v", err)
		return models.Order{}, err
	}

	return o, nil
}

func (r *implRepository) GetByID(ctx context.Context, ID string) (models.Order, error) {
	col := r.getOrderCollection()

	fil, err := r.buildGetByIDQuery(ctx, ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetByID: %v", err)
		return models.Order{}, err
	}

	var o models.Order
	if err := col.FindOne(ctx, fil).Decode(&o); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.GetByID: %v", err)
		return models.Order{}, err
	}

	return o, nil
}

func (r *implRepository) Update(ctx context.Context, ID string, opt UpdateOrderOption) (models.Order, error) {
	col := r.getOrderCollection()

	o, upDoc := r.buildUpdateOrderModel(opt)

	fil, err := r.buildGetByIDQuery(ctx, ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Update: %v", err)
		return models.Order{}, err
	}

	if _, err := col.UpdateOne(ctx, fil, upDoc); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.Update: %v", err)
		return models.Order{}, err
	}

	return o, nil
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

func (r *implRepository) List(ctx context.Context, opt ListOrderOption) ([]models.Order, error) {
	col := r.getOrderCollection()

	q := r.buildFilterQuery(ctx, opt.FilterOrder)

	cur, err := col.Find(ctx, q, options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},
		{Key: "_id", Value: -1},
	}))
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.List: %v", err)
		return nil, err
	}

	var os []models.Order
	if err := cur.All(ctx, &os); err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.List: %v", err)
		return nil, err
	}

	return os, nil
}

func (r *implRepository) GetMany(ctx context.Context, opt GetManyOrderOption) ([]models.Order, paginator.Paginator, error) {
	col := r.getOrderCollection()

	q := r.buildFilterQuery(ctx, opt.FilterOrder)

	var total int64
	os := []models.Order{}
	var wgErr error

	var wg sync.WaitGroup
	wg.Add(2)

	wg.Go(func() {
		cnt, err := col.CountDocuments(ctx, q)
		if err != nil {
			r.l.Errorf(ctx, "order.repository.mongo.GetMany.col.CountDocuments: %v", err)
			wgErr = err
			return
		}
		total = cnt
	})

	wg.Go(func() {
		cur, err := col.Find(ctx, q, options.Find().SetSkip(opt.Pag.Offset()).
			SetLimit(opt.Pag.Limit).
			SetSort(bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: -1},
			}))
		if err != nil {
			r.l.Errorf(ctx, "order.repository.mongo.GetMany.col.Find: %v", err)
			wgErr = err
			return
		}

		if err = cur.All(ctx, &os); err != nil {
			r.l.Errorf(ctx, "order.repository.mongo.GetMany.cur.All: %v", err)
			wgErr = err
			return
		}
	})

	wg.Wait()
	if wgErr != nil {
		return nil, paginator.Paginator{}, wgErr
	}

	return os, paginator.Paginator{
		Total:    total,
		Count:    int64(len(os)),
		PageSize: opt.Pag.Limit,
		Page:     opt.Pag.Page,
	}, nil

}
