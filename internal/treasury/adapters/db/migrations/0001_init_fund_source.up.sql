BEGIN;

CREATE SCHEMA IF NOT EXISTS treasury;

CREATE TYPE treasury.fund_source_type AS ENUM ('BANK', 'CASH');

CREATE TABLE treasury.fund_sources (
    fund_source_uuid        uuid                        NOT NULL,
    name                    varchar(255)                NOT NULL,
    source_type             treasury.fund_source_type   NOT NULL,
    currency                varchar(3)                  NOT NULL,

    -- balances
    balance                 DECIMAL(19, 4)              NOT NULL CHECK (balance >= 0),
    available_balance       DECIMAL(19, 4)              NOT NULL CHECK (available_balance >= 0),

    -- bank metadata
    bank_info               json,
    bank_code               varchar(20),
    bank_account_number     varchar(50),
    bank_account_owner      varchar(255),

    -- cash metadata
    cash_owner              varchar(255),

    created_at              timestamptz                 NOT NULL,
    created_by              varchar(255)                NOT NULL,
    updated_at              timestamptz,
    updated_by              varchar(255),

    PRIMARY KEY (fund_source_uuid),

    CONSTRAINT chk_bank_metadata_fields CHECK (
        (source_type = 'BANK' AND bank_account_number IS NOT NULL AND bank_account_owner IS NOT NULL AND bank_code IS NOT NULL)
        OR
        (source_type = 'CASH' AND bank_account_number IS NULL AND bank_account_owner IS NULL AND bank_code IS NULL)
    ),

    CONSTRAINT chk_cash_metadata_fields CHECK (
        (source_type = 'CASH' AND cash_owner IS NOT NULL)
        OR
        (source_type = 'BANK' AND cash_owner IS NULL)
    )
);

CREATE UNIQUE INDEX uq_fund_sources_bank_account
    ON treasury.fund_sources (bank_account_number, bank_code)
    WHERE source_type = 'BANK';

COMMIT;
