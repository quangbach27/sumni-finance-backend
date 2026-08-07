BEGIN;

CREATE TABLE treasury.wallet_allocations (
    wallet_uuid         uuid                        NOT NULL,
    fund_source_uuid    uuid                        NOT NULL,
    balance             DECIMAL(19, 4)              NOT NULL CHECK (balance >= 0),

    PRIMARY KEY (wallet_uuid, fund_source_uuid),
    CONSTRAINT fk_wallet_allocations_wallet
        FOREIGN KEY (wallet_uuid) REFERENCES treasury.wallets(wallet_uuid),
    CONSTRAINT fk_wallet_allocations_fund_source
        FOREIGN KEY (fund_source_uuid) REFERENCES treasury.fund_sources(fund_source_uuid)
);

COMMIT;
