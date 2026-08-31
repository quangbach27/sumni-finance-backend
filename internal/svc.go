package internal

import (
	"context"
	"fmt"
	"time"

	commonHTTP "sumni-finance-backend/internal/common/http"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"
	"sumni-finance-backend/internal/treasury"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type ExternalService struct{}

type Svc struct {
	echoServer *commonHTTP.EchoServer

	modules []module.Module

	dbPgx *pgxpool.Pool
}

func New(
	ctx context.Context,
	dbPgx *pgxpool.Pool,
	externalService ExternalService,
) (Svc, error) {
	server := commonHTTP.NewEchoServer()
	// TODO: Add authentication router for protectedRouter. For now, keep it with global router
	protectedRouter := server.Router

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
		err := module.RegisterHttp(ctx, server.Router, protectedRouter)
		if err != nil {
			return Svc{}, fmt.Errorf("registering http for module %s failed: %w", module.Name(), err)
		}
	}

	return Svc{
		echoServer: server,
		modules:    modules,
		dbPgx:      dbPgx,
	}, nil
}

func (s Svc) Run(ctx context.Context, port string) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := s.echoServer.Start(gCtx, port)
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		defer s.dbPgx.Close()
		return nil
	})
	return g.Wait()
}
