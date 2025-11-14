package repository

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *implRepository) buildGetByIDQuery(ctx context.Context, ID string) (bson.M, error) {
	q := bson.M{}
	q = mongo.BuildQueryWithSoftDelete(q)

	objID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.buildGetByIDQuery: %v", err)
		return nil, err
	}
	q["_id"] = objID

	return q, nil
}

func (r *implRepository) buildFilterQuery(ctx context.Context, fil order.FilterOrder) bson.M {
	q := bson.M{}
	q = mongo.BuildQueryWithSoftDelete(q)

	if fil.Code != "" {
		q["code"] = fil.Code
	}

	if fil.UserID != "" {
		q["user_id"] = fil.UserID
	}

	if fil.EventID != "" {
		q["event_id"] = fil.EventID
	}

	if fil.SessionID != "" {
		q["session_id"] = fil.SessionID
	}

	if fil.Status != nil {
		q["status"] = *fil.Status
	}

	return q
}
