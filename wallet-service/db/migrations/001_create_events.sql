-- +goose Up
CREATE TABLE events (
    global_seq      BIGSERIAL   NOT NULL,
    id              UUID        NOT NULL,
    aggregate_id    UUID        NOT NULL,
    aggregate_type  TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    event_version   INT         NOT NULL,
    payload         JSONB       NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (global_seq),
    UNIQUE (id),
    UNIQUE (aggregate_id, event_version)
);

CREATE INDEX idx_events_aggregate ON events (aggregate_id, event_version ASC);
CREATE INDEX idx_events_seq       ON events (global_seq ASC);

-- +goose Down
DROP TABLE IF EXISTS events;
