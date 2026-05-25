BEGIN;

CREATE SCHEMA IF NOT EXISTS finances;

CREATE TABLE finances.offices (
    office_uuid uuid NOT NULL,
    name varchar NOT NULL UNIQUE,
    PRIMARY KEY (office_uuid)
);

CREATE TYPE finances.fund_provider_type AS ENUM ('bank', 'cash');

CREATE TABLE finances.fund_providers 
(
    fund_provider_uuid uuid NOT NULL,
    name varchar NOT NULL UNIQUE,
    balance DECIMAL(19, 4) NOT NULL,
    available_balance DECIMAL(19, 4) NOT NULL,
    currency VARCHAR NOT NULL,
    fund_provider_type finances.fund_provider_type NOT NULL,
    bank_name varchar,
    bank_account_number varchar,
    bank_account_owner varchar,
    cash_owner_name varchar,
    office_uuid uuid NOT NULL,

    PRIMARY KEY (fund_provider_uuid),
    UNIQUE (bank_name, bank_account_number),
    FOREIGN KEY (office_uuid) REFERENCES finances.offices (office_uuid),

    CONSTRAINT chk_fund_providers_available_lte_balance  CHECK (available_balance <= balance),
    CONSTRAINT chk_fund_providers_balance_non_negative   CHECK (balance >= 0),
    CONSTRAINT chk_fund_providers_available_non_negative CHECK (available_balance >= 0)
);

CREATE TABLE finances.wallets (
    wallet_uuid uuid NOT NULL,
    name varchar NOT NULL UNIQUE,
    description varchar NOT NULL,
    balance DECIMAL(19, 4),
    currency VARCHAR NOT NULL,
    office_uuid uuid NOT NULL,

    PRIMARY KEY (wallet_uuid),
    FOREIGN KEY (office_uuid) REFERENCES finances.offices (office_uuid),

    CONSTRAINT chk_wallets_balance_non_negative   CHECK (balance >= 0)
);

CREATE TABLE finances.wallet_allocations (
    wallet_uuid uuid NOT NULL,
    fund_provider_uuid uuid NOT NULL,
    allocation_amount DECIMAL(19, 4),
    PRIMARY KEY (wallet_uuid, fund_provider_uuid),
    FOREIGN KEY (wallet_uuid) REFERENCES finances.wallets(wallet_uuid),
    FOREIGN KEY (fund_provider_uuid) REFERENCES finances.fund_providers(fund_provider_uuid)
);

COMMIT;