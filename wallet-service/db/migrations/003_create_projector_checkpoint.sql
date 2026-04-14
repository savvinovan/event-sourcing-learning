-- +goose Up
CREATE TABLE projector_checkpoints (
    projector_name  TEXT        PRIMARY KEY,
    last_global_seq BIGINT      NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed row: projector always UPDATEs, never INSERTs on the hot path.
INSERT INTO projector_checkpoints (projector_name, last_global_seq)
VALUES ('account_projector', 0);

-- +goose Down
DROP TABLE IF EXISTS projector_checkpoints;
