BEGIN;

-- 1. Create the finance schema
CREATE SCHEMA IF NOT EXISTS finance;

CREATE TABLE finance.fund_providers(
    id uuid PRIMARY KEY NOT NULL,
    name varchar(255) NOT NULL,
    fp_type varchar(255) NOT NULL,
    balance bigint NOT NULL,
    currency char(3) NOT NULL,
    unallocated_amount bigint NOT NULL,
    version int NOT NULL DEFAULT 0 
);

CREATE TABLE finance.wallets (
    id uuid PRIMARY KEY NOT NULL,
    name varchar(255) NOT NULL,
    balance bigint NOT NULL,
    currency varchar(3) NOT NULL,
    version int NOT NULL
);

CREATE TABLE finance.fund_provider_allocation(
    fund_provider_id uuid NOT NULL,
    wallet_id uuid NOT NULL,
    allocated_amount bigint NOT NULL,

    PRIMARY KEY (fund_provider_id, wallet_id),

    CONSTRAINT fk_fundprovider_allocation
        FOREIGN KEY (fund_provider_id)
            REFERENCES finance.fund_providers (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_wallet_allocation
        FOREIGN KEY (wallet_id)
            REFERENCES finance.wallets (id)
            ON DELETE CASCADE
);

COMMIT;