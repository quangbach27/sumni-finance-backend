-- name: CreateAccountingPeriod :exec
INSERT INTO finance.accounting_periods (
    id,
    year_month,
    start_date,
    interval,
    end_time,
    wallet_opening_balance,
    total_debit,
    total_credit,
    wallet_closing_balance,
    version,
    wallet_id,
    status
) VALUES (
    $1, -- id
    $2, -- year_month
    $3, -- start_date
    $4, -- interval
    $5, -- end_time
    $6, -- wallet_opening_balance
    $7, -- total_debit
    $8, -- total_credit
    $9, -- wallet_closing_balance
    $10, -- version
    $11, -- wallet_id
    $12 -- status
);

-- name: UpdateAccountingPerid :execrows
UPDATE finance.accounting_periods ap
SET
    total_debit = sqlc.arg(total_debit),
    total_credit = sqlc.arg(total_credit),
    wallet_closing_balance = sqlc.arg(closing_balance),
    status = sqlc.arg(status),
    version = version + 1
WHERE ap.id = sqlc.arg(id)
    AND ap.version = sqlc.arg(version);

-- name: GetAccountingPeriodsByIDsAndWalletID :many
SELECT
    id,
    year_month,
    start_date,
    interval,
    end_time,
    wallet_opening_balance,
    total_debit,
    total_credit,
    wallet_closing_balance,
    status,
    version
FROM
    finance.accounting_periods
WHERE wallet_id = $1 
    AND id = ANY(sqlc.arg(ids)::uuid[]);

-- name: BulkInsertTransactionRecords :copyfrom
INSERT INTO finance.transaction_records (
    id,
    transaction_no,
    transaction_type,
    amount,
    wallet_balance,
    wallet_id,
    fp_id,
    fp_balance,
    accounting_periods_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
);
