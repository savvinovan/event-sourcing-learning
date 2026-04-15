//go:build integration

package helpers

import (
	"context"

	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

// StartRedpanda starts a Redpanda (Kafka-compatible) container and returns
// the seed broker address (host:port) and a cleanup function.
// Call once in BeforeSuite; containers are shared across all specs.
func StartRedpanda(ctx context.Context, t TB) (string, func()) {
	t.Helper()

	c, err := tcredpanda.Run(ctx, "docker.redpanda.com/redpandadata/redpanda:v23.3.3")
	if err != nil {
		t.Fatalf("start redpanda container: %v", err)
	}

	broker, err := c.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("get redpanda seed broker: %v", err)
	}

	return broker, func() {
		_ = c.Terminate(ctx)
	}
}
