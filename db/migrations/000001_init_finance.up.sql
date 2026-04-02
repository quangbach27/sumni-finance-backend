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

CREATE TABLE finance.fund_provider_allocations (
    fp_id uuid NOT NULL,
    wallet_id uuid NOT NULL,
    allocated_amount bigint NOT NULL,

    PRIMARY KEY (fp_id, wallet_id),

    CONSTRAINT fk_fund_provider_allocations_fund_provider
        FOREIGN KEY (fp_id)
            REFERENCES finance.fund_providers (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_fund_provider_allocations_wallet
        FOREIGN KEY (wallet_id)
            REFERENCES finance.wallets (id)
            ON DELETE CASCADE
);

CREATE TABLE finance.accounting_periods (
    id uuid PRIMARY KEY NOT NULL,

    year_month varchar(100) NOT NULL,

    start_date int DEFAULT 1,
    interval int DEFAULT 1,

    wallet_opening_balance bigint NOT NULL,
    total_debit bigint default 0,
    total_credit bigint default 0,
    wallet_closing_balance bigint default 0,

    wallet_id uuid NOT NULL,

    CONSTRAINT fk_accounting_periods_wallet
        FOREIGN KEY (wallet_id)
            REFERENCES finance.wallets (id)
            ON DELETE CASCADE 
);

CREATE TABLE finance.transaction_records (
    id uuid PRIMARY KEY NOT NULL,
    transaction_no varchar(255),
    transaction_type varchar(10) NOT NULL,
    amount bigint NOT NULL,

    wallet_balance bigint NOT NULL,
    wallet_id uuid NOT NULL,

    fp_id uuid NOT NULL,
    fp_balance bigint NOT NULL,

    accounting_periods_id uuid NOT NULL,

    CONSTRAINT fk_transaction_record_accounting_period
        FOREIGN KEY (accounting_periods_id)
            REFERENCES finance.accounting_periods (id)
            ON DELETE CASCADE,
    
    CONSTRAINT fk_transaction_record_fund_provider
        FOREIGN KEY (fp_id)
            REFERENCES finance.fund_providers (id)
            ON DELETE CASCADE,
    
    CONSTRAINT fk_transaction_record_wallet
        FOREIGN KEY (wallet_id)
            REFERENCES finance.wallets (id)
            ON DELETE CASCADE
);

COMMIT;