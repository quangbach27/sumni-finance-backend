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
    INNER JOIN finance.fund_provider_allocation fpa
        ON fp.id = fpa.fund_provider_id
            AND fpa.wallet_id = $1;

-- name: UpdateFundProviderPartial :execrows
UPDATE finance.fund_providers
SET
    name = COALESCE(sqlc.narg(name), name),
    balance = COALESCE(sqlc.narg(balance), balance),
    unallocated_amount = COALESCE(sqlc.narg(unallocated_amount), unallocated_amount),
    currency = COALESCE(sqlc.narg(currency), currency),
    version = version + 1
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(version);

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