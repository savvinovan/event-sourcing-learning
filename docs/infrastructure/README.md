# Infrastructure Layer

## Overview

The infrastructure layer provides implementations of interfaces defined in the domain and application layers.
It contains all I/O concerns: event persistence, projections, external services.

The domain layer defines **what** must happen; infrastructure defines **how**.

## Contents

- [Event Store](eventstore.md) — `internal/infrastructure/eventstore/store.go`
