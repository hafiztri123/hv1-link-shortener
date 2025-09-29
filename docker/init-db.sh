#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_USER" <<-EOSQL
    SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '$TRANSACTION_DB') AS db_exists
    \gset
    \if :db_exists \else
        CREATE DATABASE $TRANSACTION_DB;
    \endif

    SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '$ANALYTICS_DB') AS db_exists
    \gset
    \if :db_exists \else
        CREATE DATABASE $ANALYTICS_DB;
    \endif
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$ANALYTICS_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS clicks (
        id SERIAL PRIMARY KEY,
        url_path VARCHAR(255), 
        timestamp TIMESTAMPTZ NOT NULL,
        ip_address VARCHAR(45),
        referer TEXT,
        user_agent TEXT,
        device VARCHAR(50),
        os VARCHAR(50),
        browser VARCHAR(50),
        country CHAR(2),
        city VARCHAR(255)
    );
EOSQL