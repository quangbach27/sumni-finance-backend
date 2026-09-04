BEGIN;

CREATE TABLE treasury.wallets (
    wallet_uuid         uuid                        NOT NULL,
    name                varchar(255)                NOT NULL,
    currency            varchar(3)                  NOT NULL,
    balance             DECIMAL(19, 4)              NOT NULL CHECK (balance >= 0),
    tenant_id           varchar(255)                NOT NULL,
    office_id           varchar(255)                NOT NULL,

    PRIMARY KEY (wallet_uuid)
);

COMMIT;