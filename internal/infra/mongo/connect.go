package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vogiaan1904/ticketbottle-order/config"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
)

const (
	connectTimeout = 10 * time.Second
)

// Connect connects to the database
func Connect(cfg config.MongoConfig) (mongo.Client, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), connectTimeout)
	defer cancelFunc()

	opts := mongo.NewClientOptions().
		ApplyURI(cfg.URI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to media DB: %w", err)
	}

	err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping to media DB: %w", err)
	}

	log.Println("Connected to MongoDB!")

	return client, nil
}

// Disconnect disconnects from the database.
func Disconnect(mediaClient mongo.Client) {
	if mediaClient == nil {
		return
	}

	err := mediaClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connection to MongoDB closed.")
}
