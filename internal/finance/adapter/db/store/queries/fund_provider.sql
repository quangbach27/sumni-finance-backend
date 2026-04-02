-- name: CreateFundProvider :exec
INSERT INTO finance.fund_providers (
    id,
    name,
    fp_type,
    balance,
    currency,
    unallocated_amount,
    version
) VALUES(
    $1, -- id
    $2, -- name
    $3, -- fp_type
    $4, -- balance
    $5, -- currency
    $6, -- unallocated_amount
    $7  -- version
);

-- name: GetFundProviderByWalletID :many
SELECT 
    fp.id,
    fp.name,
    fp.fp_type,
    fp.balance,
    fp.currency,
    fp.unallocated_amount,
    fp.version,
    fpa.allocated_amount AS wallet_allocated_amount
FROM finance.fund_providers fp
INNER JOIN finance.fund_provider_allocations fpa
    ON fp.id = fpa.fp_id
WHERE fpa.wallet_id = $1;

-- name: GetFundProviderByID :one
SELECT
    id,
    name,
    fp_type,
    balance,
    unallocated_amount,
    currency,
    version
FROM finance.fund_providers
WHERE id = $1;

-- name: GetFundProvidersByIDs :many
SELECT
    id,
    name,
    fp_type,
    balance,
    unallocated_amount,
    currency,
    version
FROM finance.fund_providers
WHERE id = ANY(sqlc.arg(fpIDs)::uuid[]);

-- name: BatchUpdateFundProvidersBalance :execrows
UPDATE finance.fund_providers fp
SET
    balance = v.balance,
    unallocated_amount = v.unallocated_amount,
    version = v.version + 1
FROM (
    SELECT
        unnest(sqlc.arg(ids)::uuid[]) as id,
        unnest(sqlc.arg(balances)::bigint[]) as balance,
        unnest(sqlc.arg(unallocated_amounts)::bigint[]) as unallocated_amount,
        unnest(sqlc.arg(versions)::int[]) as version
) as v
WHERE fp.id = v.id
  AND fp.version = v.version;
