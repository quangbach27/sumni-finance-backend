package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"sumni-finance-backend/internal/common/log"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lithammer/shortuuid/v3"
)

const (
	TestNameHeader          = "TestName"
	CorrelationIDHttpHeader = "Correlation-ID"
)

func rateLimiterMiddleware() echo.MiddlewareFunc {
	rps := 20.0
	if val := os.Getenv("RATE_LIMIT_RPS"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed > 0 {
			rps = parsed
		} else {
			slog.Warn("invalid RATE_LIMIT_RPS value, using default", "value", val, "default", rps)
		}
	}

	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rps,
			Burst:     int(rps) * 3,
			ExpiresIn: 3 * time.Minute,
		},
	)

	rpsHeader := fmt.Sprintf("%.0f", rps)

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: func(c *echo.Context) bool {
			return c.Request().URL.Path == "/health"
		},
		Store: store,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			ip := c.RealIP()
			if ip == "" {
				ip = "unknown"
			}
			return ip, nil
		},
		ErrorHandler: func(c *echo.Context, err error) error {
			slog.Warn("rate limiter error", "ip", c.RealIP(), "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			c.Response().Header().Set("Retry-After", "60")
			c.Response().Header().Set("X-RateLimit-Limit", rpsHeader)
			return echo.NewHTTPError(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		},
	})
}

func useMiddlewares(e *echo.Echo) {
	e.Use(
		corsMiddleware,
		rateLimiterMiddleware(),
		middleware.ContextTimeout(10*time.Second),
		middleware.Recover(),
		// Correlation-ID runs first: available in context for the request log middleware.
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				req := c.Request()
				ctx := req.Context()

				reqCorrelationID := req.Header.Get(CorrelationIDHttpHeader)
				if reqCorrelationID == "" {
					reqCorrelationID = shortuuid.New()
				}

				logger := slog.With("correlation_id", reqCorrelationID)

				if testName := c.Request().Header.Get("TestName"); testName != "" {
					logger = logger.With("test_name", testName)
				}

				ctx = log.ToContext(ctx, logger)
				ctx = log.ContextWithCorrelationID(ctx, reqCorrelationID)
				c.SetRequest(req.WithContext(ctx))
				c.Response().Header().Set(CorrelationIDHttpHeader, reqCorrelationID)

				return next(c)
			}
		},
		requestLogMiddleware,
	)
}

func corsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	allowedOrigins := []string{"*"}

	if originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS"); originsEnv != "" {
		origins := strings.Split(originsEnv, ";")
		allowedOrigins = make([]string, 0, len(origins))

		for _, origin := range origins {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	corsConfig := middleware.CORSConfig{
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			CorrelationIDHttpHeader,
			TestNameHeader,
		},
		ExposeHeaders:    []string{CorrelationIDHttpHeader},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		// Echo v5 refuses AllowOrigins=["*"] combined with AllowCredentials=true (insecure combination
		// per the CORS spec). UnsafeAllowOriginFunc reflects the request's own Origin back, which is
		// wire-compatible with the old AllowOrigins=["*"] behavior while satisfying that guard.
		corsConfig.UnsafeAllowOriginFunc = func(_ *echo.Context, origin string) (string, bool, error) {
			return origin, true, nil
		}
	} else {
		corsConfig.AllowOrigins = allowedOrigins
	}

	return middleware.CORSWithConfig(corsConfig)(next)
}

type bodyCapturingWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyCapturingWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyCapturingWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyCapturingWriter) Flush() {
	err := http.NewResponseController(w.ResponseWriter).Flush()
	if err != nil && !errors.Is(err, http.ErrNotSupported) {
		slog.Warn("response writer flush failed", "error", err)
	}
}

func (w *bodyCapturingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *bodyCapturingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func requestLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Read request body and restore it for the handler.
		var reqBody []byte
		if c.Request().Body != nil {
			reqBody, _ = io.ReadAll(c.Request().Body)
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// Capture response body via MultiWriter.
		resBody := new(bytes.Buffer)
		respWriter := c.Response()
		mw := io.MultiWriter(respWriter, resBody)
		c.SetResponse(&bodyCapturingWriter{Writer: mw, ResponseWriter: respWriter})

		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		ctx := c.Request().Context()

		status := http.StatusOK
		if echoResp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
			status = echoResp.Status
		}

		logger := log.FromContext(ctx).With(
			"URI", c.Request().RequestURI,
			"status", status,
			"method", c.Request().Method,
			"duration", duration.String(),
		)
		if err != nil {
			logger = logger.With("error", err)
		}
		logger = logger.With("request_body", truncateBodyForLog(string(reqBody)))

		body := resBody.String()
		if utf8.ValidString(body) {
			if isDebug := log.FromContext(ctx).Enabled(ctx, slog.LevelDebug); !isDebug {
				body = truncateBodyForLog(body)
			}
			logger = logger.With("response_body", body)
		} else {
			logger = logger.With("response_body", "<binary data>")
		}

		logger.Info("Request done")
		return err
	}
}
