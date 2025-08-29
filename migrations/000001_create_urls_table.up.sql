CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(20) UNIQUE,
    long_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
