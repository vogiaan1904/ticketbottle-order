package repository

import (
	"time"

	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
)

type implRepository struct {
	l     logger.Logger
	db    mongo.Database
	clock func() time.Time
}

var _ Repository = &implRepository{}

func New(l logger.Logger, db mongo.Database) Repository {
	return &implRepository{
		l:     l,
		db:    db,
		clock: time.Now,
	}
}
