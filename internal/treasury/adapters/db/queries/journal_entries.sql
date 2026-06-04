-- name: BatchInsertJournalEntries :copyfrom
INSERT INTO treasury.journal_entries (
    journal_entry_uuid,
    fund_source_uuid,
    amount,
    entry_type,
    balance_before,
    balance_after,
    status,
    transaction_date,
    transaction_no,
    description,
    created_at,
    created_by
) VALUES (
    sqlc.arg(journal_entry_uuid),
    sqlc.arg(fund_source_uuid),
    sqlc.arg(amount),
    sqlc.arg(entry_type),
    sqlc.arg(balance_before),
    sqlc.arg(balance_after),
    sqlc.arg(status),
    sqlc.arg(transaction_date),
    sqlc.arg(transaction_no),
    sqlc.arg(description),
    sqlc.arg(created_at),
    sqlc.arg(created_by)
);

-- name: UpdateJournalEntry :exec
UPDATE treasury.journal_entries
SET
    status             = sqlc.arg(status),
    reverse_entry_uuid = sqlc.narg(reverse_entry_uuid),
    voided_by          = sqlc.narg(voided_by),
    voided_at          = sqlc.narg(voided_at),
    voided_reason      = sqlc.narg(voided_reason),
    updated_at         = sqlc.arg(updated_at),
    updated_by         = sqlc.arg(updated_by)
WHERE journal_entry_uuid = sqlc.arg(journal_entry_uuid);

-- name: GetJournalEntryByUID :one
SELECT *
FROM treasury.journal_entries
WHERE journal_entry_uuid = $1;