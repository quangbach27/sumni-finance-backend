package treasury

import (
	"context"
	"embed"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"

	"sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/api/http"
	"sumni-finance-backend/internal/treasury/app/command"
	"sumni-finance-backend/internal/treasury/app/query"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pgxDb *pgxpool.Pool

	commandHandler     *command.Handlers
	queryHandler       *query.Handlers
	bankLookupProvider domain.BankLookupProvider
}

func NewModule(pgxDb *pgxpool.Pool, bankLookupProvider domain.BankLookupProvider) *Module {
	if pgxDb == nil {
		panic("db can't be nil")
	}

	if bankLookupProvider == nil {
		panic("bank lookup provider can't be empty")
	}

	return &Module{
		pgxDb:              pgxDb,
		bankLookupProvider: bankLookupProvider,
	}
}

func (m *Module) Name() module.Name {
	return "treasury"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	if err := common.MigrateDatabaseUp(
		ctx,
		string(m.Name()),
		m.pgxDb,
		embedMigrations,
		"adapters/db/migrations",
	); err != nil {
		return err
	}

	fundSourceRepo := db.NewFundSourceRepository(m.pgxDb)
	fundSouceFactory := domain.NewFundSourceFactory(m.bankLookupProvider)
	fundSourceReadModel := db.NewFundSourceReadModel(m.pgxDb)
	journalEntryReadModel := db.NewJournalEntryReadModel(m.pgxDb)

	m.commandHandler = command.NewHandlers(fundSourceRepo, fundSouceFactory)
	m.queryHandler = query.NewHandlers(fundSourceReadModel, journalEntryReadModel)

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	return nil
}

func (m *Module) RegisterHttp(
	ctx context.Context,
	publicRouter common.EchoRouter,
	protectedRouter common.EchoRouter,
) error {
	return http.Register(protectedRouter, m.commandHandler, m.queryHandler)
}
