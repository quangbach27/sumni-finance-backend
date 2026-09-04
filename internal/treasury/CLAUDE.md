# CLAUDE.md — treasury module

This file provides guidance to Claude Code (claude.ai/code) when working inside `internal/treasury/`. See the root
`CLAUDE.md` for repo-wide commands, module wiring, and layout conventions.

## Domain model: fund sources vs. wallets

`FundSource` (`domain/fund_source.go`) is the **physical** money source — a real BANK account or CASH tin, with its
own `balance` and `Metadata` (`BankMetadata`/`CashMetadata`, matched to `FundSourceType`). It tracks two numbers:
`balance` (the real total) and `availableBalance` (`balance` minus whatever has been reserved against it). `TopUp`/
`Withdraw` move the real `balance`; `Reserve` only carves into `availableBalance` and never touches `balance` — it's
how a wallet claims a slice of the source without moving real money yet.

`Wallet` (`domain/wallet.go`) is a **virtual** layer on top of one or more fund sources: `Wallet.Allocate(fs, amount)`
calls `fs.Reserve(amount)` and records a `FundSourceAllocation{fundSource, balance}` in `Wallet.allocations`. A
wallet's own `balance` is just the sum of its allocations, and `Wallet.TopUp`/`Withdraw` operate through a specific
allocation's `FundSourceUUID`, mutating both the allocation and the underlying fund source's real balance together.
So one fund source can back multiple wallets (each holding a different reserved slice of its `availableBalance`),
and a wallet never holds money on its own — it's a view that manages the fund sources linked to it. `Transaction`
records are the ledger entries produced by these moves (see lifecycle below).

## Transaction lifecycle (internal/treasury/domain/transaction.go)

Transactions move through `DRAFTED -> RECORDED -> POSTED -> VOIDED -> REVERSED` (state machine documented in a
comment on `TransactionStatus`). `Void()` doesn't mutate balances itself — it produces a paired reversing
`Transaction` with flipped `EntryType` and marks the original `VOIDED`; the caller is responsible for persisting
both. Follow this pattern (produce a companion transaction rather than mutating amounts) for any future changes to
posted transactions.
