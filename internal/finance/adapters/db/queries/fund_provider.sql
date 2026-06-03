-- name: FundProvidersByUUIDs :many
SELECT *
FROM finances.fund_providers
WHERE fund_provider_uuid = ANY(sqlc.arg(fund_provider_uuids)::UUID[]);

-- name: FundProviderByUUID :one
SELECT *
FROM finances.fund_providers
WHERE fund_provider_uuid = $1;

-- name: InsertFundProvider :exec
INSERT INTO finances.fund_providers (
    fund_provider_uuid,
    name,
    balance,
    available_balance,
    currency,
    fund_provider_type,
    bank_info,
    bank_account_number,
    bank_account_owner,
    cash_owner_name,
    office_uuid
)
VALUES (
    sqlc.arg(fund_provider_uuid),
    sqlc.arg(name),
    sqlc.arg(balance),
    sqlc.arg(available_balance),
    sqlc.arg(currency),
    sqlc.arg(fund_provider_type),
    sqlc.arg(bank_info),
    sqlc.arg(bank_account_number),
    sqlc.arg(bank_account_owner),
    sqlc.arg(cash_owner_name),
    sqlc.arg(office_uuid)
);

-- name: UpdateFundProviderBalance :exec
UPDATE finances.fund_providers
SET 
    balance = COALESCE(sqlc.narg(balance), balance),
    available_balance =  COALESCE(sqlc.narg(available_balance), available_balance)
WHERE
    fund_provider_uuid = sqlc.arg(fund_provider_uuid)
;
