package app

import (
	"sumni-finance-backend/internal/common/cqrs"
	"sumni-finance-backend/internal/finance/adapter/db"
	"sumni-finance-backend/internal/finance/adapter/db/store"
	"sumni-finance-backend/internal/finance/app/command"

	common_db "sumni-finance-backend/internal/common/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	AllocateFund             command.AllocateFundHandler
	CreateFundProvider       command.CreateFundProviderHandler
	CreateWallet             command.CreateWalletHandler
	OpenAccountingPeriod     command.OpenAccountingPeriodHandler
	RecordTransactionRecords command.RecordTransactionRecordsHandler
}

type Queries struct {
}

func NewApplication(pgPool *pgxpool.Pool) (Application, error) {
	queries := store.New(pgPool)
	transactionManager := common_db.NewPgxTransactionManager(pgPool)

	walletRepo, err := db.NewWalletRepo(queries, transactionManager)
	if err != nil {
		return Application{}, err
	}

	fundProviderRepo, err := db.NewFundProviderRepo(queries)
	if err != nil {
		return Application{}, err
	}

	ledgerRepo := db.NewLedgerRepository(queries)

	return Application{
		Commands: Commands{
			AllocateFund:             cqrs.ApplyCommandDecorators(command.NewAllocateFundHandler(walletRepo, fundProviderRepo)),
			CreateFundProvider:       cqrs.ApplyCommandDecorators(command.NewCreateFundProviderHandler(fundProviderRepo)),
			CreateWallet:             cqrs.ApplyCommandDecorators(command.NewCreateWalletHandler(walletRepo)),
			OpenAccountingPeriod:     cqrs.ApplyCommandDecorators(command.NewOpenAccountingPeriodHandler(walletRepo, ledgerRepo)),
			RecordTransactionRecords: cqrs.ApplyCommandDecorators(command.NewRecordTransactionRecordsHandler(walletRepo)),
		},
		Queries: Queries{},
	}, nil
}
