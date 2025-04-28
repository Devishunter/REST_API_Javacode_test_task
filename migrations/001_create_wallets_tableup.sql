-- migrations/001_create_wallets_tableup.sql
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0
);