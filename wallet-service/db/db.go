// Package db embeds SQL migrations for use with goose at startup.
package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
