BEGIN;

CREATE TYPE treasury.transaction_status AS ENUM ('DRAFTED', 'RECORDED', 'POSTED', 'VOIDED', 'REVERSED');
CREATE TYPE treasury.entry_type AS ENUM ('in', 'out');

CREATE TABLE treasury.transactions (
    transaction_uuid            uuid                          NOT NULL,
    status                      treasury.transaction_status  NOT NULL,
    entry_type                  treasury.entry_type           NOT NULL,
    amount                      DECIMAL(19, 4)                NOT NULL CHECK (amount > 0),
    currency                    varchar(3)                    NOT NULL,
    description                 text,
    transaction_date            timestamptz                   NOT NULL,

    tenant_id                   varchar(255)                NOT NULL,
    office_id                   varchar(255)                NOT NULL,

    fund_source_uuid            uuid                          NOT NULL,
    wallet_uuid                 uuid,
    reversed_transaction_uuid   uuid,

    PRIMARY KEY (transaction_uuid),

    FOREIGN KEY (fund_source_uuid) REFERENCES treasury.fund_sources(fund_source_uuid),
    FOREIGN KEY (wallet_uuid) REFERENCES treasury.wallets(wallet_uuid),
    FOREIGN KEY (reversed_transaction_uuid) REFERENCES treasury.transactions(transaction_uuid)
);

COMMIT;
