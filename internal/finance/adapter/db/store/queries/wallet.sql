-- name: CreateWallet :exec
INSERT INTO finance.wallets (
    id,
    name,
    balance,
    currency,
    version
) VALUES (
    $1, -- id
    $2, -- name
    $3, -- balance
    $4, -- currency
    $5 -- version
);

-- name: GetWalletByID :one
SELECT 
    id,
    name,
    balance,
    currency,
    version
FROM finance.wallets
WHERE id = $1;

-- name: UpdateWalletBalance :execrows
UPDATE finance.wallets
SET
    balance = sqlc.arg(balance),
    version = version + 1
WHERE id = sqlc.arg(id)
    AND version = sqlc.arg(version);

-- name: BulkInsertFundAllocations :copyfrom
INSERT INTO finance.fund_provider_allocations (
    fp_id,
    wallet_id,
    allocated_amount
) VALUES ($1, $2, $3);
