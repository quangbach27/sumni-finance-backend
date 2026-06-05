BEGIN;

CREATE TYPE treasury.journal_entry_type AS ENUM ('DEBIT', 'CREDIT');
CREATE TYPE treasury.journal_entry_status AS ENUM ('RECORDED', 'POSTED', 'VOIDED', 'REVERSE');

CREATE TABLE treasury.journal_entries (
    journal_entry_uuid      uuid                            NOT NULL,
    fund_source_uuid        uuid                            NOT NULL,
    amount                  DECIMAL(19, 4)                  NOT NULL CHECK (amount > 0),
    entry_type              treasury.journal_entry_type     NOT NULL,
    balance_before          DECIMAL(19, 4)                  NOT NULL,
    balance_after           DECIMAL(19, 4)                  NOT NULL,
    status                  treasury.journal_entry_status   NOT NULL,
    transaction_date        timestamptz                     NOT NULL,
    transaction_no          varchar(255),
    description             text,

    reverse_entry_uuid      uuid,
    voided_by               varchar(255),
    voided_at               timestamptz,
    voided_reason           text,

    created_at              timestamptz                     NOT NULL,
    created_by              varchar(255)                    NOT NULL,
    updated_at              timestamptz,
    updated_by              varchar(255),

    PRIMARY KEY (journal_entry_uuid),

    FOREIGN KEY (fund_source_uuid)
        REFERENCES treasury.fund_sources(fund_source_uuid),

    FOREIGN KEY (reverse_entry_uuid)
        REFERENCES treasury.journal_entries(journal_entry_uuid),

    CONSTRAINT chk_balance_after CHECK (
        (entry_type = 'DEBIT'  AND balance_after = balance_before - amount) OR
        (entry_type = 'CREDIT' AND balance_after = balance_before + amount)
    ),

    CONSTRAINT chk_void_fields CHECK (
        (status = 'VOIDED'
            AND voided_by IS NOT NULL
            AND voided_at IS NOT NULL
            AND voided_reason IS NOT NULL
            AND reverse_entry_uuid IS NOT NULL)
        OR
        (status != 'VOIDED'
            AND voided_by IS NULL
            AND voided_at IS NULL
            AND voided_reason IS NULL)
    )
);

CREATE UNIQUE INDEX uq_journal_entries_transaction_no
    ON treasury.journal_entries (transaction_no)
    WHERE status != 'VOIDED';

CREATE INDEX idx_journal_entries_fund_source_uuid
    ON treasury.journal_entries (fund_source_uuid);

COMMIT;