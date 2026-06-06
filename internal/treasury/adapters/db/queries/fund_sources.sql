-- name: InsertFundSource :exec
INSERT INTO treasury.fund_sources (
    fund_source_uuid,
    name,
    source_type,
    balance,
    available_balance,
    currency,
    bank_info,
    bank_code,
    bank_account_number,
    bank_account_owner,
    cash_owner,
    created_at,
    created_by
) VALUES (
    sqlc.arg(fund_source_uuid),
    sqlc.arg(name),
    sqlc.arg(source_type),
    sqlc.arg(balance),
    sqlc.arg(available_balance),
    sqlc.arg(currency),
    sqlc.arg(bank_info),
    sqlc.arg(bank_code),
    sqlc.arg(bank_account_number),
    sqlc.arg(bank_account_owner),
    sqlc.arg(cash_owner),
    sqlc.arg(created_at),
    sqlc.arg(created_by)
);

-- name: UpdateFundSource :exec
UPDATE treasury.fund_sources
SET
    balance           = sqlc.arg(balance),
    available_balance = sqlc.arg(available_balance),
    updated_at        = sqlc.arg(updated_at),
    updated_by        = sqlc.arg(updated_by)
WHERE fund_source_uuid = sqlc.arg(fund_source_uuid);

-- name: GetFundSourceByUUID :one
SELECT *
FROM treasury.fund_sources
WHERE fund_source_uuid = sqlc.arg(fund_source_uuid);

-- name: ListFundSources :many
SELECT *
FROM treasury.fund_sources;