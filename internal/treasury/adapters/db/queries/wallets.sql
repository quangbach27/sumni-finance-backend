-- name: InsertWallet :exec
INSERT INTO treasury.wallets (
    wallet_uuid,
    name,
    currency,
    balance,
    tenant_id,
    office_id
) VALUES (
    sqlc.arg(wallet_uuid),
    sqlc.arg(name),
    sqlc.arg(currency),
    sqlc.arg(balance),
    sqlc.arg(tenant_id),
    sqlc.arg(office_id)
);

-- name: GetWalletByUUID :one
SELECT *
FROM treasury.wallets
WHERE wallet_uuid = sqlc.arg(wallet_uuid)
    AND tenant_id = sqlc.arg(tenant_id)
    AND office_id = sqlc.arg(office_id);
