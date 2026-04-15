//go:build integration

package helpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// TestKafkaConsumer reads messages from a single Kafka partition, starting from
// the offset snapshotted at creation time. Any message published after creation
// is guaranteed to be returned by ReadMessageInto.
type TestKafkaConsumer struct {
	conn  *kafka.Conn
	topic string
}

// NewTestKafkaConsumer creates a consumer seeked to the current end of the partition.
// Call BEFORE the action that publishes; read AFTER.
func NewTestKafkaConsumer(ctx context.Context, t TB, broker, topic string) *TestKafkaConsumer {
	t.Helper()

	conn, err := kafka.DialLeader(ctx, "tcp", broker, topic, 0)
	if err != nil {
		t.Fatalf("dial kafka leader for %s: %v", topic, err)
	}

	lastOffset, err := conn.ReadLastOffset()
	if err != nil {
		t.Fatalf("read last offset for %s: %v", topic, err)
	}
	if _, err := conn.Seek(lastOffset, kafka.SeekAbsolute); err != nil {
		t.Fatalf("seek to end of %s: %v", topic, err)
	}

	return &TestKafkaConsumer{conn: conn, topic: topic}
}

// ReadMessageInto reads the next message and unmarshals its value into dst.
func (c *TestKafkaConsumer) ReadMessageInto(t TB, timeout time.Duration, dst any) {
	t.Helper()

	if err := c.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatalf("set deadline on kafka conn: %v", err)
	}

	batch := c.conn.ReadBatch(1, 1<<20)
	defer func() { _ = batch.Close() }()

	msg, err := batch.ReadMessage()
	if err != nil {
		t.Fatalf("read message from %s (timeout %s): %v", c.topic, timeout, err)
	}

	if err := json.Unmarshal(msg.Value, dst); err != nil {
		t.Fatalf("unmarshal message from %s: %v", c.topic, err)
	}
}

func (c *TestKafkaConsumer) Close() {
	_ = c.conn.Close()
}
