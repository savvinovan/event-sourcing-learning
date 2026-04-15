//go:build integration

package helpers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TruncateAll resets all event-sourcing tables and re-seeds the projector checkpoint.
// Call in BeforeEach to isolate each spec from all others.
func TruncateAll(ctx context.Context, t TB, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(ctx,
		`TRUNCATE TABLE events, account_read_models, transaction_history RESTART IDENTITY CASCADE`,
	)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	// Reset projector checkpoint to 0 — the row always exists (seeded by migration).
	_, err = pool.Exec(ctx,
		`UPDATE projector_checkpoints SET last_global_seq = 0, updated_at = now() WHERE projector_name = 'account_projector'`,
	)
	if err != nil {
		t.Fatalf("reset projector checkpoint: %v", err)
	}
}
