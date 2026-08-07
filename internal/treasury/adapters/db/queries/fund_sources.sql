-- name: InsertFundSource :exec
INSERT INTO treasury.fund_sources (
    fund_source_uuid,
    name,
    source_type,
    balance,
    available_balance,
    currency,
    bank_info,
    bin,
    bank_account_number,
    bank_account_owner,
    cash_owner,
    tenant_id,
    office_id
) VALUES (
    sqlc.arg(fund_source_uuid),
    sqlc.arg(name),
    sqlc.arg(source_type),
    sqlc.arg(balance),
    sqlc.arg(available_balance),
    sqlc.arg(currency),
    sqlc.arg(bank_info),
    sqlc.arg(bin),
    sqlc.arg(bank_account_number),
    sqlc.arg(bank_account_owner),
    sqlc.arg(cash_owner),
    sqlc.arg(tenant_id),
    sqlc.arg(office_id)
);

-- name: GetFundSourceByUUID :one
SELECT *
FROM treasury.fund_sources
WHERE fund_source_uuid = sqlc.arg(fund_source_uuid);

-- name: GetFundSourcesByUUIDs :many
SELECT *
FROM treasury.fund_sources
WHERE fund_source_uuid = ANY(sqlc.arg(fund_source_uuids)::uuid[]) 
  AND tenant_id = sqlc.arg(tenant_id)
  AND office_id = sqlc.arg(office_id);

-- name: UpdateFundSourceAvailableBalance :exec
UPDATE treasury.fund_sources
SET available_balance = sqlc.arg(new_available_balance)
WHERE fund_source_uuid = sqlc.arg(fund_source_uuid);
