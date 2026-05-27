-- name: WalletByUUID :one
SELECT *
FROM finances.wallets
WHERE wallet_uuid = $1;

-- name: InsertWallet :exec
INSERT INTO finances.wallets (
    wallet_uuid,
    name,
    description,
    balance,
    currency,
    office_uuid
)
VALUES (
    sqlc.arg(wallet_uuid),
    sqlc.arg(name),
    sqlc.arg(description),
    sqlc.arg(balance),
    sqlc.arg(currency),
    sqlc.arg(office_uuid)
);

-- name: UpsertFundProviderAllocations :exec
INSERT INTO finances.fund_provider_allocations (
    wallet_uuid,
    fund_provider_uuid,
    allocation_amount
)
VALUES ($1, $2, $3)
ON CONFLICT (wallet_uuid, fund_provider_uuid) 
    DO UPDATE SET
        allocation_amount = EXCLUDED.allocation_amount;

-- name: FundProvidersWithAllocations :many
SELECT 
    sqlc.embed(fpa),
    sqlc.embed(fp)
FROM finances.fund_provider_allocations fpa
    INNER JOIN finances.fund_providers fp ON fpa.fund_provider_uuid = fp.fund_provider_uuid
WHERE fpa.wallet_uuid = sqlc.arg(wallet_uuid);