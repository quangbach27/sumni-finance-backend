package http

import (
	"log/slog"
	"net/http"
	"time"

	"sumni-finance-backend/internal/common"

	"github.com/labstack/echo/v4"
)

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	useMiddlewares(e)

	e.HTTPErrorHandler = common.EchoErrorHandler
	e.Logger = common.NewEchoSlogAdapter(slog.Default())

	e.Server.WriteTimeout = 30 * time.Second
	e.Server.ReadHeaderTimeout = 30 * time.Second
	e.Server.ReadTimeout = 30 * time.Second
	e.Server.IdleTimeout = 60 * time.Second

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}
