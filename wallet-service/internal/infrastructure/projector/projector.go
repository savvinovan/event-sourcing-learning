// Package projector implements the async projection pattern:
// a separate process tails the events table and maintains read model tables.
package projector

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/savvinovan/wallet-service/internal/domain/event"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
)

// EventApplier updates read model tables for a batch of domain events.
// Called inside the projector's transaction — if it returns an error,
// the transaction rolls back and the checkpoint is not advanced.
type EventApplier interface {
	Apply(ctx context.Context, tx pgx.Tx, events []event.DomainEvent) error
}

// Runner polls the events table for new events above the last checkpoint,
// applies them to the read model, and advances the checkpoint — all within
// a single database transaction per batch.
type Runner struct {
	db            *pgxpool.Pool
	registry      *eventstore.Registry
	upcasters     *eventstore.UpcasterRegistry
	applier       EventApplier
	projectorName string
	batchSize     int
	pollInterval  time.Duration
	log           *slog.Logger
}

func NewRunner(
	db *pgxpool.Pool,
	registry *eventstore.Registry,
	upcasters *eventstore.UpcasterRegistry,
	applier EventApplier,
	projectorName string,
	batchSize int,
	pollInterval time.Duration,
	log *slog.Logger,
) *Runner {
	return &Runner{
		db:            db,
		registry:      registry,
		upcasters:     upcasters,
		applier:       applier,
		projectorName: projectorName,
		batchSize:     batchSize,
		pollInterval:  pollInterval,
		log:           log,
	}
}

// Run starts the event loop. Blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	r.log.Info("projector started", "name", r.projectorName, "batch_size", r.batchSize, "poll_interval", r.pollInterval)
	for {
		n, err := r.processBatch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // shutdown
			}
			r.log.Error("projector batch failed", "error", err)
			// back off briefly on error to avoid tight error loops
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(r.pollInterval):
			}
			continue
		}

		if n == 0 {
			// No new events — wait before polling again.
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(r.pollInterval):
			}
		}
		// If batch was full, immediately poll again (catch-up mode).
	}
}

// processBatch fetches up to batchSize events above the checkpoint, applies them,
// and advances the checkpoint — all in one transaction.
// Returns the number of events processed.
func (r *Runner) processBatch(ctx context.Context) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("projector: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock the checkpoint row so only one projector instance processes at a time.
	var checkpoint int64
	err = tx.QueryRow(ctx,
		`SELECT last_global_seq FROM projector_checkpoints WHERE projector_name = $1 FOR UPDATE`,
		r.projectorName,
	).Scan(&checkpoint)
	if err != nil {
		return 0, fmt.Errorf("projector: read checkpoint: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT global_seq, aggregate_id, aggregate_type, event_type, event_version, schema_version, payload, occurred_at
		FROM events
		WHERE global_seq > $1
		ORDER BY global_seq ASC
		LIMIT $2
	`, checkpoint, r.batchSize)
	if err != nil {
		return 0, fmt.Errorf("projector: fetch events: %w", err)
	}
	defer rows.Close()

	var (
		domainEvents []event.DomainEvent
		maxSeq       int64
	)
	for rows.Next() {
		var (
			seq           int64
			aggregateID   string
			aggType       string
			eventType     string
			version       int
			schemaVersion int
			payload       []byte
			occurredAt    time.Time
		)
		if err := rows.Scan(&seq, &aggregateID, &aggType, &eventType, &version, &schemaVersion, &payload, &occurredAt); err != nil {
			return 0, fmt.Errorf("projector: scan event row: %w", err)
		}
		payload, err = r.upcasters.Upcast(eventType, schemaVersion, payload)
		if err != nil {
			return 0, fmt.Errorf("projector: upcast seq=%d: %w", seq, err)
		}
		base := event.RestoreBase(aggregateID, aggType, eventType, version, occurredAt)
		e, err := r.registry.Deserialize(eventType, base, payload)
		if err != nil {
			return 0, fmt.Errorf("projector: deserialize seq=%d: %w", seq, err)
		}
		domainEvents = append(domainEvents, e)
		maxSeq = seq
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("projector: rows error: %w", err)
	}
	if len(domainEvents) == 0 {
		return 0, nil
	}

	if err := r.applier.Apply(ctx, tx, domainEvents); err != nil {
		return 0, fmt.Errorf("projector: apply: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE projector_checkpoints SET last_global_seq = $1, updated_at = now() WHERE projector_name = $2`,
		maxSeq, r.projectorName,
	)
	if err != nil {
		return 0, fmt.Errorf("projector: update checkpoint: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("projector: commit: %w", err)
	}

	r.log.Debug("projector batch committed", "count", len(domainEvents), "up_to_seq", maxSeq)
	return len(domainEvents), nil
}
