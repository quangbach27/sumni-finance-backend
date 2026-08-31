package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"sumni-finance-backend/internal/common"

	"github.com/labstack/echo/v5"
)

type EchoServer struct {
	Router *echo.Echo
}

func NewEchoServer() *EchoServer {
	e := echo.New()

	useMiddlewares(e)

	e.HTTPErrorHandler = common.EchoErrorHandler
	e.Logger = slog.Default()

	e.GET("/health", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return &EchoServer{
		Router: e,
	}
}

func (es *EchoServer) Start(ctx context.Context, port string) error {
	startConfig := echo.StartConfig{
		Address:         port,
		HideBanner:      true,
		GracefulTimeout: 30 * time.Second,
		BeforeServeFunc: func(srv *http.Server) error {
			srv.IdleTimeout = 60 * time.Second
			srv.ReadHeaderTimeout = 30 * time.Second
			return nil
		},
	}

	if err := startConfig.Start(ctx, es.Router); err != nil {
		return fmt.Errorf("starting http server failed: %w", err)
	}

	return nil
}
