package repository

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *implRepository) buildGetByIDQuery(ctx context.Context, ID string) (bson.M, error) {
	fil := bson.M{}
	fil = mongo.BuildQueryWithSoftDelete(fil)

	objID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		r.l.Errorf(ctx, "order.repository.OrderRepository.buildGetByIDQuery: %v", err)
		return nil, err
	}
	fil["_id"] = objID

	return fil, nil
}

func (r *implRepository) buildGetByCodeQuery(ctx context.Context, code string) bson.M {
	fil := bson.M{}
	fil = mongo.BuildQueryWithSoftDelete(fil)

	fil["code"] = code

	return fil
}
