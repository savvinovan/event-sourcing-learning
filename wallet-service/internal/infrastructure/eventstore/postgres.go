package eventstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/savvinovan/wallet-service/internal/domain/event"
)

const pgUniqueViolation = "23505"

// PostgresEventStore persists domain events in a PostgreSQL events table.
// Optimistic concurrency is enforced by both an application-level version check
// and the UNIQUE (aggregate_id, event_version) DB constraint.
type PostgresEventStore struct {
	db       *pgxpool.Pool
	registry *Registry
}

func NewPostgresEventStore(db *pgxpool.Pool, registry *Registry) *PostgresEventStore {
	return &PostgresEventStore{db: db, registry: registry}
}

// Append inserts new events for an aggregate stream.
// Returns ErrVersionConflict if the current stream version differs from expectedVersion,
// or if a concurrent writer inserted an event with the same version (unique constraint).
func (s *PostgresEventStore) Append(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres event store: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Application-level optimistic concurrency check (fast path).
	var currentVersion int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(event_version), 0) FROM events WHERE aggregate_id = $1::uuid`,
		aggregateID,
	).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("postgres event store: check version: %w", err)
	}
	if currentVersion != expectedVersion {
		return ErrVersionConflict
	}

	for _, e := range events {
		payload, err := s.registry.Serialize(e)
		if err != nil {
			return fmt.Errorf("postgres event store: serialize %s: %w", e.EventType(), err)
		}

		eventID := uuid.Must(uuid.NewV7()).String()

		_, err = tx.Exec(ctx, `
			INSERT INTO events (id, aggregate_id, aggregate_type, event_type, event_version, payload, occurred_at)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
		`, eventID, e.AggregateID(), e.AggregateType(), e.EventType(), e.Version(), payload, e.OccurredAt())
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				return ErrVersionConflict
			}
			return fmt.Errorf("postgres event store: insert event: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres event store: commit: %w", err)
	}
	return nil
}

// Load returns all events for an aggregate in ascending version order.
func (s *PostgresEventStore) Load(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT aggregate_id::text, aggregate_type, event_type, event_version, payload, occurred_at
		FROM events
		WHERE aggregate_id = $1::uuid
		ORDER BY event_version ASC
	`, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("postgres event store: load: %w", err)
	}
	defer rows.Close()

	return scanEvents(rows, s.registry)
}

func scanEvents(rows pgx.Rows, registry *Registry) ([]event.DomainEvent, error) {
	var result []event.DomainEvent
	for rows.Next() {
		var (
			aggregateID string
			aggType     string
			eventType   string
			version     int
			payload     []byte
			occurredAt  time.Time
		)
		if err := rows.Scan(&aggregateID, &aggType, &eventType, &version, &payload, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}
		base := event.RestoreBase(aggregateID, aggType, eventType, version, occurredAt)
		e, err := registry.Deserialize(eventType, base, payload)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

var _ EventStore = (*PostgresEventStore)(nil)
