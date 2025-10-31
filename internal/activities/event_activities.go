package activities

import (
	"context"
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
)

type EventActivities struct {
	Client event.EventServiceClient
}

func NewEventActivities(client event.EventServiceClient) *EventActivities {
	return &EventActivities{
		Client: client,
	}
}

func (a *EventActivities) GetEvent(ctx context.Context, eventID string) (*event.Event, error) {
	resp, err := a.Client.FindOne(ctx, &event.FindOneEventRequest{
		Id: eventID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if resp.Event == nil {
		return nil, fmt.Errorf("event not found")
	}

	return resp.Event, nil
}

func (a *EventActivities) GetEventConfig(ctx context.Context, eventID string) (*event.EventConfig, error) {
	resp, err := a.Client.GetConfig(ctx, &event.GetEventConfigRequest{
		EventId: eventID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get event config: %w", err)
	}

	if resp.EventConfig == nil {
		return nil, fmt.Errorf("event config not found")
	}

	return resp.EventConfig, nil
}
