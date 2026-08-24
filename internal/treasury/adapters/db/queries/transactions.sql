-- name: InsertTransaction :exec
INSERT INTO treasury.transactions (
    transaction_uuid,
    status,
    entry_type,
    amount,
    currency,
    description,
    transaction_date,
    tenant_id,
    office_id,
    fund_source_uuid,
    wallet_uuid,
    reversed_transaction_uuid
) VALUES (
    @transaction_uuid,
    @status,
    @entry_type,
    @amount,
    @currency,
    @description,
    @transaction_date,
    @tenant_id,
    @office_id,
    @fund_source_uuid,
    @wallet_uuid,
    @reversed_transaction_uuid
);

-- name: GetTransactionByUUID :one
SELECT *
FROM treasury.transactions
WHERE transaction_uuid = @transaction_uuid;
