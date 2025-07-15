CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE IF NOT EXISTS pairs(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    average_price DECIMAL(10,2),
    min_price DECIMAL(10,2),
    max_price DECIMAL(10,2)
);

CREATE INDEX IF NOT EXISTS idx_pairs_name ON pairs(name);
CREATE INDEX IF NOT EXISTS idx_pairs_exchange ON pairs(exchange);
CREATE INDEX IF NOT EXISTS idx_pairs_time ON pairs(time);
CREATE INDEX IF NOT EXISTS idx_pairs_name_exchange_time ON pairs(name, exchange, time DESC);
