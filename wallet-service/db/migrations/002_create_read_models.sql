-- +goose Up
CREATE TABLE account_read_models (
    account_id    UUID              PRIMARY KEY,
    customer_id   UUID              NOT NULL,
    status        TEXT              NOT NULL,
    balance       NUMERIC(20, 8)    NOT NULL DEFAULT 0,
    currency      TEXT              NOT NULL,
    version       INT               NOT NULL,
    updated_at    TIMESTAMPTZ       NOT NULL
);

CREATE TABLE transaction_history (
    id            UUID              PRIMARY KEY,
    account_id    UUID              NOT NULL,
    tx_type       TEXT              NOT NULL,   -- 'deposit' | 'withdrawal'
    amount        NUMERIC(20, 8)    NOT NULL,
    currency      TEXT              NOT NULL,
    occurred_at   TIMESTAMPTZ       NOT NULL
);

CREATE INDEX idx_tx_account ON transaction_history (account_id, occurred_at ASC);

-- +goose Down
DROP TABLE IF EXISTS transaction_history;
DROP TABLE IF EXISTS account_read_models;
