CREATE TABLE IF NOT EXISTS promocode_usage (
    id SERIAL PRIMARY KEY,
    promocode_id BIGINT REFERENCES promocode(id),
    used_by BIGINT NOT NULL REFERENCES customer(telegram_id),
    used_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
