package identity

import (
	"context"
	"embed"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"
	"sumni-finance-backend/internal/identity/api/http"
	"sumni-finance-backend/internal/identity/app"
	"sumni-finance-backend/internal/identity/app/models"

	identityModule "sumni-finance-backend/internal/identity/api/module"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	config *common.Config
	dbPgx  *pgxpool.Pool

	handler http.Handlers
	service *app.Service

	authenticator  http.Authenticator
	sessionStore   models.SessionStore
	policyEnforcer app.PolicyEnforcer
}

func NewModule(
	config *common.Config,
	dbPgx *pgxpool.Pool,
	authenticator http.Authenticator,
	sessionStore models.SessionStore,
	policyEnforcer app.PolicyEnforcer,
) *Module {
	if dbPgx == nil {
		panic("pgxDb can't be nil")
	}

	if authenticator == nil {
		panic("auth can't be nil")
	}

	if sessionStore == nil {
		panic("session store can't be nil")
	}

	if policyEnforcer == nil {
		panic("policyEnforcer can't be nil")
	}

	return &Module{
		config:         config,
		dbPgx:          dbPgx,
		authenticator:  authenticator,
		sessionStore:   sessionStore,
		policyEnforcer: policyEnforcer,
	}
}

func (m *Module) Name() module.Name {
	return "identities"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	if err := common.MigrateDatabaseUp(
		ctx,
		string(m.Name()),
		m.dbPgx,
		embedMigrations,
		"adapters/db/migrations",
	); err != nil {
		return err
	}

	m.service = app.NewService(m.policyEnforcer)
	m.handler = http.NewHandlers(m.config, m.authenticator, m.sessionStore)

	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	contracts.Identity = identityModule.New(m.service)
	return nil
}

func (m *Module) RegisterHttp(
	ctx context.Context,
	publicRouter common.EchoRouter,
	protectedRouter common.EchoRouter,
) error {
	return http.Register(
		m.config,
		publicRouter,
		protectedRouter,
		m.handler,
	)
}
