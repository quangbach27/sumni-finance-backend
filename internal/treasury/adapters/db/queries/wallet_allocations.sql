-- name: CreateWalletAllocations :copyfrom
INSERT INTO treasury.wallet_allocations (
    wallet_uuid,
    fund_source_uuid,
    balance
) VALUES ($1, $2, $3);

-- name: GetAllocationsByWalletUUID :many
SELECT sqlc.embed(wa), sqlc.embed(fs)
FROM treasury.wallet_allocations wa
    JOIN treasury.fund_sources fs ON fs.fund_source_uuid = wa.fund_source_uuid
WHERE wa.wallet_uuid = sqlc.arg(wallet_uuid);