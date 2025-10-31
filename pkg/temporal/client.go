package temporal

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

type Config struct {
	HostPort  string
	Namespace string
}

// NewClient creates a new Temporal client
func NewClient(cfg Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return c, nil
}
