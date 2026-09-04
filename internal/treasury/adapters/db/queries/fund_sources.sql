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
    @fund_source_uuid,
    @name,
    @source_type,
    @balance,
    @available_balance,
    @currency,
    @bank_info,
    @bin,
    @bank_account_number,
    @bank_account_owner,
    @cash_owner,
    @tenant_id,
    @office_id
);

-- name: GetFundSourceByUUID :one
SELECT *
FROM treasury.fund_sources
WHERE fund_source_uuid = @fund_source_uuid;

-- name: GetFundSourcesByUUIDs :many
SELECT *
FROM treasury.fund_sources
WHERE fund_source_uuid = ANY(@fund_source_uuids::uuid[])
  AND tenant_id = @tenant_id
  AND office_id = @office_id;

-- name: UpdateFundSourceAvailableBalance :exec
UPDATE treasury.fund_sources
SET available_balance = @new_available_balance
WHERE fund_source_uuid = @fund_source_uuid;

-- name: UpdateFundSourceBalance :exec
UPDATE treasury.fund_sources
SET balance = @new_balance
WHERE fund_source_uuid = @fund_source_uuid;

-- name: ListFundSources :many
SELECT *
FROM treasury.fund_sources
WHERE tenant_id = @tenant_id
  AND office_id = @office_id
ORDER BY fund_source_uuid ASC
LIMIT @limit_val
OFFSET @offset_val;

-- name: CountFundSources :one
SELECT COUNT(*)
FROM treasury.fund_sources
WHERE tenant_id = @tenant_id
  AND office_id = @office_id;
