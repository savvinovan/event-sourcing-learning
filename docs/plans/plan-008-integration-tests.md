# PLAN-008: Integration Tests with Testcontainers

| | |
|-|-|
| **Status** | DONE |
| **Date** | 2026-04-13 |
| **Updated** | 2026-04-15 |
| **Depends on** | [PLAN-007](plan-007-postgresql-event-store.md), [PLAN-006](plan-006-event-driven-integration.md) |

## Goal

Integration tests that spin up real PostgreSQL and Kafka in Docker via `testcontainers-go`.
Tests verify the full stack — from HTTP handler through command/query handlers, event store, projector, and Kafka broker.
No mocks for infrastructure dependencies.

**Test stack:** Ginkgo v2 + Gomega, testcontainers-go, YAML fixtures.

## Why This Stack

| Choice | Reason |
|--------|--------|
| **Ginkgo v2** | `Describe`/`Context`/`It` maps cleanly to DDD ubiquitous language; `BeforeSuite`/`AfterSuite` for containers; `BeforeEach` for DB cleanup |
| **Gomega** | Rich matchers (`HaveOccurred`, `ConsistOf`, `Eventually`, `Satisfy`) — especially `Eventually` for async Kafka consumer assertions |
| **testcontainers-go** | Real containers, no mock/prod divergence; official Postgres + Redpanda modules |
| **Redpanda** | Kafka-compatible, ~3x faster startup than full Kafka — no ZooKeeper, single binary |
| **YAML fixtures** | Scenario data outside code; easy to read, diff, and add new cases without recompiling |
| **goleak** | Detect goroutine leaks from unclosed consumers or pooled connections |

## Directory Structure

```
wallet-service/
  test/
    integration/
      suite_test.go              # Ginkgo bootstrap, BeforeSuite/AfterSuite
      account_test.go            # account flow specs
      concurrent_withdrawal_test.go
      kafka_consumer_test.go     # KYCVerified → account activated
      testdata/
        fixtures/
          open_account.yaml
          deposit_withdraw.yaml
          kyc_verified.yaml
          concurrent_withdrawals.yaml
      helpers/
        containers.go            # PostgreSQL + Redpanda container helpers
        fixtures.go              # YAML fixture loader
        db.go                    # DB truncate / seed helpers
        kafka.go                 # test Kafka producer helper

kyc-service/
  test/
    integration/
      suite_test.go
      kyc_test.go                # KYC flow specs
      kafka_publisher_test.go    # KYCVerified / KYCRejected published to Kafka
      testdata/
        fixtures/
          submit_kyc.yaml
          approve_kyc.yaml
          reject_kyc.yaml
      helpers/
        containers.go
        fixtures.go
        kafka.go

test/                            # cross-service (go.work workspace level)
  integration/
    suite_test.go
    full_flow_test.go            # submit KYC → approve → wallet activated → withdraw
    helpers/
      ...
```

## Dependencies to Add

### wallet-service and kyc-service `go.mod`

```go
require (
    github.com/onsi/ginkgo/v2          v2.22.x
    github.com/onsi/gomega             v1.36.x
    github.com/testcontainers/testcontainers-go         v0.37.x
    github.com/testcontainers/testcontainers-go/modules/postgres v0.37.x
    github.com/testcontainers/testcontainers-go/modules/redpanda v0.37.x
    go.uber.org/goleak                 v1.3.x
    gopkg.in/yaml.v3                   v3.0.1   // already indirect → promote to direct
)
```

> Ginkgo + Gomega only in test binaries; add to `_test.go` imports. No production code dependency.

## Container Helpers

### `helpers/containers.go`

```go
//go:build integration

package helpers

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
    tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

// StartPostgres starts a PostgreSQL container, runs migrations via goose,
// and returns a ready pool + cleanup func.
func StartPostgres(ctx context.Context, t testing.TB, migrations embed.FS) (*pgxpool.Pool, string, func()) {
    c, err := tcpostgres.Run(ctx,
        "postgres:16-alpine",
        tcpostgres.WithDatabase("wallet_test"),
        tcpostgres.WithUsername("wallet"),
        tcpostgres.WithPassword("wallet"),
        tcpostgres.WithInitScripts(), // goose applied separately
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(30*time.Second),
        ),
    )
    // ... run goose migrations, return pool + dsn + cleanup
}

// StartRedpanda starts a Redpanda (Kafka-compatible) container.
// Returns bootstrap brokers string + cleanup func.
func StartRedpanda(ctx context.Context, t testing.TB) (string, func()) {
    c, err := tcredpanda.Run(ctx, "docker.redpanda.com/redpandadata/redpanda:latest")
    // ... return brokers + cleanup
}
```

**Key pattern:** containers start once in `BeforeSuite`, shared across all specs in a suite.
Each spec gets a clean DB via `BeforeEach` truncate — not a new container.

## Suite Setup Pattern

```go
//go:build integration

package integration_test

import (
    "context"
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "go.uber.org/goleak"

    "github.com/savvinovan/wallet-service/test/integration/helpers"
)

func TestIntegration(t *testing.T) {
    // goleak: fail if goroutines leak after suite
    defer goleak.VerifyNone(t,
        goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
    )
    RegisterFailHandler(Fail)
    RunSpecs(t, "Wallet Service Integration Suite")
}

var (
    pool        *pgxpool.Pool
    brokers     string
    cleanupPG   func()
    cleanupRP   func()
)

var _ = BeforeSuite(func() {
    ctx := context.Background()
    pool, _, cleanupPG = helpers.StartPostgres(ctx, GinkgoT(), db.Migrations)
    brokers, cleanupRP  = helpers.StartRedpanda(ctx, GinkgoT())
})

var _ = AfterSuite(func() {
    cleanupPG()
    cleanupRP()
})
```

## Test Isolation: BeforeEach Truncate

```go
var _ = BeforeEach(func() {
    _, err := pool.Exec(context.Background(),
        `TRUNCATE TABLE events, account_read_models, projector_checkpoints RESTART IDENTITY CASCADE`)
    Expect(err).NotTo(HaveOccurred())
})
```

One clean slate per spec. No shared state between `It` blocks.

## YAML Fixtures

### Format

```yaml
# testdata/fixtures/open_account.yaml
scenario: "open account happy path"
input:
  currency: "USD"
expected:
  status: "Pending"
  balance: "0.00"
  events:
    - type: "AccountOpened"
      fields:
        currency: "USD"

---
scenario: "open account — duplicate ID is idempotent"
input:
  currency: "EUR"
expected:
  status: "Pending"
  balance: "0.00"
```

### Loader

```go
//go:build integration

package helpers

import (
    "embed"
    "gopkg.in/yaml.v3"
)

type Fixture[T any] struct {
    Scenario string `yaml:"scenario"`
    Input    T      `yaml:"input"`
    Expected any    `yaml:"expected"`
}

// LoadFixtures decodes all YAML documents in a fixture file into a slice.
func LoadFixtures[T any](fs embed.FS, path string) []Fixture[T] {
    data, err := fs.ReadFile(path)
    // ... yaml.NewDecoder loop over multi-doc YAML
}
```

### Usage in Spec

```go
//go:embed testdata/fixtures/open_account.yaml
var openAccountFixtures embed.FS

var _ = Describe("Opening an account", func() {
    fixtures := helpers.LoadFixtures[OpenAccountInput](openAccountFixtures, "testdata/fixtures/open_account.yaml")

    for _, f := range fixtures {
        f := f // capture
        It(f.Scenario, func() {
            cmd := account.OpenAccountCommand{
                AccountID:  domain.NewAccountID(),
                CustomerID: domain.NewCustomerID(),
                Currency:   f.Input.Currency,
            }
            err := cmdBus.Dispatch(ctx, cmd)
            Expect(err).NotTo(HaveOccurred())

            // verify read model
            acc, err := readRepo.GetByID(ctx, cmd.AccountID)
            Expect(err).NotTo(HaveOccurred())
            Expect(string(acc.Status)).To(Equal(f.Expected.Status))
            Expect(acc.Balance.String()).To(Equal(f.Expected.Balance))
        })
    }
})
```

## Async Assertions with `Eventually`

For Kafka consumer tests (projector, kyc consumer), use Gomega `Eventually`:

```go
It("activates account when KYCVerified received", func() {
    // arrange: open account
    // act: publish KYCVerified to Kafka topic
    helpers.PublishKYCVerified(ctx, brokers, accountID)

    // assert: Eventually the read model is updated (consumer is async)
    Eventually(func(g Gomega) {
        acc, err := readRepo.GetByID(ctx, accountID)
        g.Expect(err).NotTo(HaveOccurred())
        g.Expect(string(acc.Status)).To(Equal("Active"))
    }).WithTimeout(10 * time.Second).WithPolling(200 * time.Millisecond).Should(Succeed())
})
```

## Concurrent Withdrawal Test

```go
Describe("Concurrent withdrawals", func() {
    It("exactly one succeeds, one returns ErrVersionConflict", func() {
        // setup: funded account
        // fire two concurrent withdrawals
        var errs [2]error
        var wg sync.WaitGroup
        wg.Add(2)
        for i := 0; i < 2; i++ {
            go func() {
                defer wg.Done()
                errs[i] = cmdBus.Dispatch(ctx, withdrawCmd)
            }()
        }
        wg.Wait()

        successes := lo.CountBy(errs[:], func(e error) bool { return e == nil })
        conflicts := lo.CountBy(errs[:], func(e error) bool {
            return errors.Is(e, domain.ErrVersionConflict)
        })
        Expect(successes).To(Equal(1))
        Expect(conflicts).To(Equal(1))
    })
})
```

## Test Coverage

### wallet-service

| Spec | Fixture file | Assertion |
|------|-------------|-----------|
| Open account → events in PG | `open_account.yaml` | read model status = Pending |
| Deposit → balance updated | `deposit_withdraw.yaml` | read model balance = input amount |
| Withdraw → balance updated | `deposit_withdraw.yaml` | read model balance decremented |
| Withdraw insufficient funds | `deposit_withdraw.yaml` | `ErrInsufficientFunds` returned |
| Concurrent withdrawals | `concurrent_withdrawals.yaml` | exactly 1 success, 1 failure (`ErrVersionConflict` or `ErrInsufficientFunds`) |
| KYCVerified from Kafka → Active | `kyc_verified.yaml` | `Eventually` account.Status = Active |

### kyc-service

| Spec | Fixture file | Assertion |
|------|-------------|-----------|
| Submit KYC → event persisted | `submit_kyc.yaml` | event in store |
| Approve KYC → KYCVerified on Kafka | `approve_kyc.yaml` | Kafka message consumed |
| Reject KYC → KYCRejected on Kafka | `reject_kyc.yaml` | Kafka message consumed |

### Cross-service (covered within wallet-service)

The workspace-level `test/` module was removed because Go prohibits importing `internal/` packages
across module boundaries. Cross-service scenarios are instead covered in wallet-service integration
tests via a test Kafka producer that publishes raw contract events.

| Spec | Assertion |
|------|-----------|
| KYCVerified from Kafka → account Active | `Eventually` account.Status = Active |
| KYCRejected from Kafka → account Frozen | `Eventually` account.Status = Frozen |

## Build Tags and Makefile

```go
//go:build integration
```

All integration test files carry this tag. Unit tests have no tag.

```makefile
.PHONY: test test-integration

test:
	go test ./...

test-integration:
	go test -tags=integration -timeout=5m -v ./test/integration/...

test-all:
	go test -race ./...
	go test -tags=integration -race -timeout=5m ./test/integration/...
```

## CI: GitHub Actions

```yaml
# .github/workflows/integration.yml
name: Integration Tests
on:
  push:
    branches: [main]
  pull_request:

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: wallet-service/go.mod
      - name: Run integration tests
        run: make test-integration
        env:
          TESTCONTAINERS_RYUK_DISABLED: "false"  # keep Ryuk for CI cleanup
```

> Testcontainers uses the Docker socket on GitHub Actions runners — no extra setup needed.

## Acceptance Criteria

- [ ] `go test -tags=integration ./...` runs with no pre-installed PostgreSQL or Kafka on the host
- [ ] Each spec starts with a clean DB — BeforeEach truncates all tables
- [ ] Concurrent withdrawal spec reliably produces exactly 1 success and 1 `ErrVersionConflict`
- [ ] Async Kafka assertions use `Eventually` — no `time.Sleep`
- [ ] YAML fixtures load from embedded FS — no filesystem path dependencies
- [ ] goleak passes after suite — no goroutine leaks
- [ ] Cross-service full-flow spec passes end-to-end
- [ ] All integration tests pass in GitHub Actions CI
- [ ] Suite run time under 2 minutes on CI

## Tasks

- [x] Add Ginkgo v2, Gomega, testcontainers-go, goleak to both service `go.mod`
- [x] `helpers/containers.go` — PostgreSQL container: start, run goose migrations, return pool + cleanup
- [x] `helpers/containers.go` — Redpanda container: start, return brokers + cleanup
- [x] `helpers/fixtures.go` — generic YAML fixture loader (multi-doc, embed.FS)
- [x] `helpers/db.go` — TruncateAll helper for BeforeEach
- [x] `helpers/kafka.go` — test Kafka producer helper (publish arbitrary contract events)
- [x] wallet-service: `suite_test.go` with BeforeSuite/AfterSuite/goleak
- [x] wallet-service: YAML fixture files in `testdata/fixtures/`
- [x] wallet-service: `account_test.go` — open, deposit, withdraw specs
- [x] wallet-service: `concurrent_withdrawal_test.go` — concurrency spec
- [x] wallet-service: `kafka_consumer_test.go` — KYCVerified → Active with `Eventually`
- [x] kyc-service: `suite_test.go`, fixture files, `kyc_test.go`, `kafka_publisher_test.go`
- [x] Cross-service: workspace-level `test/integration/full_flow_test.go`
- [x] Makefile targets: `test-integration`, `test-all`
- [x] CI: `.github/workflows/integration.yml`
- [x] Update docs
