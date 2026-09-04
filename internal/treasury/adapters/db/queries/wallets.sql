-- name: InsertWallet :exec
INSERT INTO treasury.wallets (
    wallet_uuid,
    name,
    currency,
    balance,
    tenant_id,
    office_id
) VALUES (
    @wallet_uuid,
    @name,
    @currency,
    @balance,
    @tenant_id,
    @office_id
);

-- name: GetWalletByUUID :one
SELECT *
FROM treasury.wallets
WHERE wallet_uuid = @wallet_uuid
    AND tenant_id = @tenant_id
    AND office_id = @office_id;

-- name: ListWallets :many
SELECT *
FROM treasury.wallets
WHERE tenant_id = @tenant_id
  AND office_id = @office_id
ORDER BY wallet_uuid ASC
LIMIT @limit_val
OFFSET @offset_val;

-- name: CountWallets :one
SELECT COUNT(*)
FROM treasury.wallets
WHERE tenant_id = @tenant_id
  AND office_id = @office_id;

-- name: UpdateWalletBalance :exec
UPDATE treasury.wallets
SET balance = $1
WHERE wallet_uuid = $2;

-- name: GetWalletsWithAllocationsByUUIDs :many
WITH requested_pairs AS (
    SELECT unnest(@wallet_uuids::uuid[]) AS wallet_uuid,
           unnest(@fund_source_uuids::uuid[]) AS fund_source_uuid
)
SELECT sqlc.embed(w), sqlc.embed(wa), sqlc.embed(fs)
FROM treasury.wallets w
JOIN requested_pairs rp ON rp.wallet_uuid = w.wallet_uuid
JOIN treasury.wallet_allocations wa
    ON wa.wallet_uuid = rp.wallet_uuid AND wa.fund_source_uuid = rp.fund_source_uuid
JOIN treasury.fund_sources fs ON fs.fund_source_uuid = wa.fund_source_uuid
WHERE w.tenant_id = @tenant_id
  AND w.office_id = @office_id;
