-- +goose Up
CREATE TABLE products_views
(
    event_time DateTime64(3, 'UTC') DEFAULT now64(3),

    product_id UInt64,
    user_id Nullable(UInt64),
    session_id String,

    event_id UUID DEFAULT generateUUIDv4(),

    source LowCardinality(String) DEFAULT 'unknown',
    device LowCardinality(String) DEFAULT 'unknown',
    country_code FixedString(2) DEFAULT '--'
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (product_id, event_time, event_id);


-- +goose Down
DROP TABLE IF EXISTS products_views;