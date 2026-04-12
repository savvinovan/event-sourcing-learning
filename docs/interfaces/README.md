# Interface Adapters

## Overview

Interface adapters translate external requests (HTTP, gRPC, CLI) into application-layer commands and queries,
and translate results back into external response formats.

They are allowed to depend on the application layer but must not contain business logic.

## Contents

- [HTTP API](http.md) — `internal/interfaces/http/`
