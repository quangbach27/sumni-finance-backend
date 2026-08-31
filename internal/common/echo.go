package common

import (
	"github.com/labstack/echo/v5"
)

type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
}
