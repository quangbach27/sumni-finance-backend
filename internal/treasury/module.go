package treasury

import (
	"context"
	"embed"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"

	"sumni-finance-backend/internal/treasury/app/command"
	"sumni-finance-backend/internal/treasury/app/query"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pgxDb *pgxpool.Pool

	commandHandler *command.Handlers
	queryHandler   *query.Handlers
}

func NewModule(pgxDb *pgxpool.Pool) *Module {
	if pgxDb == nil {
		panic("db can't be nil")
	}

	return &Module{
		pgxDb: pgxDb,
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

	m.commandHandler = command.NewHandlers()
	m.queryHandler = query.NewHandlers()

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return nil
}
