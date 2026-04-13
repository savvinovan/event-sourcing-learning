# PLAN-008: Integration Tests with Testcontainers

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-007](plan-007-postgresql-event-store.md), [PLAN-006](plan-006-event-driven-integration.md) |

## Goal

Integration tests that spin up real PostgreSQL and Kafka in Docker via `testcontainers-go`.
Tests verify the full stack — from HTTP request through command handler, event store, and event broker.
No mocks for infrastructure dependencies.

## Why Testcontainers

- Tests run against the real DB and real Kafka — no mock/prod divergence
- Each test suite gets a fresh container — no shared state between test runs
- Works in CI without pre-installed infrastructure

## Test Coverage

### wallet-service
- Open account → verify events persisted in PostgreSQL
- Deposit / Withdraw → verify balance projection updated
- Concurrent withdrawals → verify optimistic concurrency conflict is handled
- Receive `KYCVerified` from Kafka → verify account activated

### kyc-service
- Submit KYC → verify event persisted
- Approve KYC → verify `KYCVerified` published to Kafka topic
- Reject KYC → verify `KYCRejected` published to Kafka topic

### Cross-service
- Full flow: submit KYC → approve → wallet activated → withdrawal succeeds

## Acceptance Criteria

- [ ] `go test ./... -tags integration` runs with no pre-installed PostgreSQL or Kafka on the host
- [ ] Each test suite starts with a clean DB — no state leaks between tests
- [ ] Concurrent withdrawal test reliably produces exactly one success and one `ErrVersionConflict`
- [ ] Cross-service test passes: open account → submit KYC → approve → account becomes Active → withdraw succeeds
- [ ] All integration tests pass in GitHub Actions CI
- [ ] Test run time is under 2 minutes on CI

## Tasks

- [ ] Add `testcontainers-go` to both services
- [ ] PostgreSQL test helper: start container, run migrations, return connection string
- [ ] Kafka test helper: start container, return bootstrap servers
- [ ] wallet-service integration tests (see coverage above)
- [ ] kyc-service integration tests (see coverage above)
- [ ] Cross-service integration test (full flow)
- [ ] CI: run integration tests in GitHub Actions
- [ ] Update docs
