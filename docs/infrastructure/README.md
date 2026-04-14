# Infrastructure Layer

## Overview

The infrastructure layer provides implementations of interfaces defined in the domain and application layers.
It contains all I/O concerns: event persistence, projections, external services.

The domain layer defines **what** must happen; infrastructure defines **how**.

## Contents

- [Event Store](eventstore.md) — `wallet-service/internal/infrastructure/eventstore/`
- [Async Projector](projector.md) — `wallet-service/internal/infrastructure/projector/` and `wallet-service/cmd/projector/`
