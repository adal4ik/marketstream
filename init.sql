CREATE TABLE pairs(
    pair_id SERIAL PRIMARY KEY,
    pair_name VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    average_price DECIMAL(10,2),
    min_price DECIMAL(10,2),
    max_price DECIMAL(10,2)
);