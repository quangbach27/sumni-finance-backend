BEGIN;

-- 1. Drop dependent tables first (in reverse order of foreign key dependencies)
DROP TABLE IF EXISTS finance.transaction_records;
DROP TABLE IF EXISTS finance.accounting_periods;
DROP TABLE IF EXISTS finance.fund_provider_allocations;

-- 2. Drop base tables
DROP TABLE IF EXISTS finance.wallets;
DROP TABLE IF EXISTS finance.fund_providers;

-- 3. Drop schema (only if empty)
DROP SCHEMA IF EXISTS finance;

COMMIT;