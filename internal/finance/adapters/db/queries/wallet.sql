-- name: WalletByUUID :one
SELECT *
FROM finances.wallets
WHERE wallet_uuid = $1;