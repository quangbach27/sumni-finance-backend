package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	commonHTTP "sumni-finance-backend/internal/common/http"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"
	"sumni-finance-backend/internal/treasury"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type ExternalService struct{}

type Svc struct {
	echoRouter *echo.Echo

	modules []module.Module

	dbPgx *pgxpool.Pool
}

func New(
	ctx context.Context,
	dbPgx *pgxpool.Pool,
	externalService ExternalService,
) (Svc, error) {
	e := commonHTTP.NewEcho()

	moduleContracts := &contracts.Contracts{}

	modules := []module.Module{
		treasury.NewModule(dbPgx),
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
		err := module.RegisterHttp(ctx, e)
		if err != nil {
			return Svc{}, fmt.Errorf("registering http for module %s failed: %w", module.Name(), err)
		}
	}

	return Svc{
		echoRouter: e,
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

	s.echoRouter.Server.WriteTimeout = 30 * time.Second
	s.echoRouter.Server.ReadHeaderTimeout = 30 * time.Second
	s.echoRouter.Server.ReadTimeout = 30 * time.Second
	s.echoRouter.Server.IdleTimeout = 60 * time.Second

	err := s.echoRouter.Start(port)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("starting http server failed: %w", err)
	}

	return nil
}
