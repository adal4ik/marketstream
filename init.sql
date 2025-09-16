CREATE TABLE IF NOT EXISTS minute_aggregates (
    pair_name     TEXT        NOT NULL,
    exchange      TEXT        NOT NULL,
    "timestamp"   TIMESTAMPTZ NOT NULL,
    average_price DOUBLE PRECISION NOT NULL,
    min_price     DOUBLE PRECISION NOT NULL,
    max_price     DOUBLE PRECISION NOT NULL,
    CONSTRAINT uq_minute UNIQUE (pair_name, exchange, "timestamp")
);

CREATE INDEX IF NOT EXISTS idx_min_agg_ts
    ON minute_aggregates ("timestamp");

CREATE INDEX IF NOT EXISTS idx_min_agg_pair_ex_ts
    ON minute_aggregates (pair_name, exchange, "timestamp" DESC);
