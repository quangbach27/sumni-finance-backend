package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sumni-finance-backend/internal/common"
	commonHTTP "sumni-finance-backend/internal/common/http"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"
	"sumni-finance-backend/internal/identity"
	identityDb "sumni-finance-backend/internal/identity/adapters/db"
	identityHTTP "sumni-finance-backend/internal/identity/api/http"
	identityModel "sumni-finance-backend/internal/identity/app/models"
	"sumni-finance-backend/internal/treasury"
	treasuryDomain "sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type ExternalService struct {
	BankLookupProvider treasuryDomain.BankLookupProvider
	Authenticator      identityHTTP.Authenticator
	SessionManager     identityModel.SessionManager
}

type Svc struct {
	echoRouter *echo.Echo
	config     *common.Config

	modules []module.Module

	dbPgx *pgxpool.Pool
}

func New(
	ctx context.Context,
	config *common.Config,
	dbPgx *pgxpool.Pool,
	externalService ExternalService,
) (Svc, error) {
	sessionStore := identityDb.NewSessionRepository(dbPgx)
	authMiddleware := identityHTTP.AuthMiddleware(sessionStore, externalService.SessionManager)

	rootRouter := commonHTTP.NewEcho(config)
	protectedRouter := rootRouter.Group("", authMiddleware)

	moduleContracts := &contracts.Contracts{}

	modules := []module.Module{
		identity.NewModule(config, dbPgx, externalService.Authenticator, sessionStore),
		treasury.NewModule(dbPgx, externalService.BankLookupProvider),
	}

	for _, module := range modules {
		start := time.Now()

		if err := module.Init(ctx); err != nil {
			return Svc{}, fmt.Errorf("initializing module %s failed: %w", module.Name(), err)
		}

		if err := module.RegisterContracts(ctx, moduleContracts); err != nil {
			return Svc{}, fmt.Errorf("failed to register contracts: %w", err)
		}

		log.FromContext(ctx).With(
			"duration", time.Since(start),
			"module", module.Name(),
		).Debug("Initialized module")
	}

	if err := moduleContracts.Verify(); err != nil {
		return Svc{}, fmt.Errorf("verifying module contracts failed: %w", err)
	}

	for _, module := range modules {
		err := module.RegisterHttp(ctx, rootRouter, protectedRouter)
		if err != nil {
			return Svc{}, fmt.Errorf("registering http for module %s failed: %w", module.Name(), err)
		}
	}

	return Svc{
		echoRouter: rootRouter,
		config:     config,
		modules:    modules,
		dbPgx:      dbPgx,
	}, nil
}

func (s Svc) Run(ctx context.Context, port string) error {
	defer s.dbPgx.Close()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := s.echoRouter.Shutdown(shutdownCtx)
		if err != nil {
			log.FromContext(ctx).Error("shutting down http server failed")
		}
	}()

	if s.config.App.Env != "development" {
		s.echoRouter.Server.WriteTimeout = 30 * time.Second
		s.echoRouter.Server.ReadHeaderTimeout = 30 * time.Second
		s.echoRouter.Server.ReadTimeout = 30 * time.Second
		s.echoRouter.Server.IdleTimeout = 60 * time.Second
	}

	err := s.echoRouter.Start(port)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("starting http server failed: %w", err)
	}

	return nil
}
