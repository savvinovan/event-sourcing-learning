# Interface Adapters

## Overview

Interface adapters translate external requests (HTTP, gRPC, CLI) into application-layer commands and queries,
and translate results back into external response formats.

They are allowed to depend on the application layer but must not contain business logic.

## Contents

- [Wallet HTTP API](http.md) — `wallet-service/internal/interfaces/http/`
- [KYC HTTP API](kyc-http.md) — `kyc-service/internal/interfaces/http/`
