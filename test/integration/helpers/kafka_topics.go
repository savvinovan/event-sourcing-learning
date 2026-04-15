//go:build integration

package helpers

import (
	"context"
	"fmt"
	"net"

	"github.com/segmentio/kafka-go"
)

// CreateTopics creates Kafka topics on the given broker.
// Call once in BeforeSuite, before starting any consumers or producers.
func CreateTopics(ctx context.Context, t TB, broker string, topics ...string) {
	t.Helper()

	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("dial kafka broker %s: %v", broker, err)
	}
	defer conn.Close() //nolint:errcheck

	ctrl, err := conn.Controller()
	if err != nil {
		t.Fatalf("get kafka controller: %v", err)
	}

	ctrlAddr := fmt.Sprintf("%s:%d", ctrl.Host, ctrl.Port)
	ctrlConn, err := kafka.DialContext(ctx, "tcp", ctrlAddr)
	if err != nil {
		var dialErr error
		ctrlConn, dialErr = kafka.DialContext(ctx, "tcp", broker)
		if dialErr != nil {
			t.Fatalf("dial kafka controller %s: %v (original broker error: %v)", ctrlAddr, err, dialErr)
		}
	}
	defer ctrlConn.Close() //nolint:errcheck

	configs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		configs[i] = kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}

	if err := ctrlConn.CreateTopics(configs...); err != nil {
		if isTopicAlreadyExistsErr(err) {
			return
		}
		t.Fatalf("create kafka topics: %v", err)
	}
}

func isTopicAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	var kerr kafka.Error
	if ok := asKafkaError(err, &kerr); ok {
		return kerr == kafka.TopicAlreadyExists
	}
	return false
}

func asKafkaError(err error, target *kafka.Error) bool {
	if _, ok := err.(net.Error); ok {
		return false
	}
	if e, ok := err.(kafka.Error); ok {
		*target = e
		return true
	}
	return false
}
