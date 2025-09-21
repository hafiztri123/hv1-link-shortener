#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'app_db') AS db_exists
    \gset
    \if :db_exists \else
        CREATE DATABASE app_db;
    \endif

    SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'analytics_db') AS db_exists
    \gset
    \if :db_exists \else
        CREATE DATABASE analytics_db;
    \endif
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "analytics_db" <<-EOSQL
    CREATE TABLE IF NOT EXISTS clicks (
        id SERIAL PRIMARY KEY,
        timestamp TIMESTAMPTZ NOT NULL,
        ip_address VARCHAR(45),
        referrer TEXT,
        user_agent TEXT,
        device VARCHAR(50),
        os VARCHAR(50),
        browser VARCHAR(50),
        country CHAR(2),
        city VARCHAR(255)
    );
EOSQL