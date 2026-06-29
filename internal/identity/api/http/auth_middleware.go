package http

import (
	"context"
	"errors"
	"net/http"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/identity/app/models"

	"github.com/labstack/echo/v4"
)

type ctxKey string

const sessionKey ctxKey = "session"

func AuthMiddleware(sessionStore models.SessionStore, manager models.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := c.Request().Context()
			logger := log.FromContext(ctx)

			sessionID, err := getSessionIDFromCookie(c)
			if err != nil {
				return common.NewUnauthorizedError(
					"missing-session-id",
					"session id is not exist in cookie").
					WithInternalError(err)
			}

			session, err := sessionStore.GetActiveSessionByID(ctx, sessionID)
			if err != nil {
				if errors.Is(err, models.ErrActiveSessionNotFound) {
					return common.NewUnauthorizedError("session-not-found", "session not found").WithInternalError(err)
				}

				return common.NewInternalServerError("internal-error", "server is error").WithInternalError(err)
			}

			verifyErr := manager.VerifySession(ctx, session)
			if err == nil {
				ctx = context.WithValue(ctx, sessionKey, session)
				c.SetRequest(req.WithContext(ctx))
				logger.Info("session verified")

				return next(c)
			}

			logger.Info("session expired. Rotating session",
				"verify_session_error", verifyErr,
			)

			newSession, rotateErr := rotateSession(
				ctx,
				manager,
				sessionStore,
				session,
			)
			if rotateErr != nil {
				return rotateErr
			}

			c.SetCookie(&http.Cookie{
				Name:     sessionCookieName,
				Value:    newSession.ID.String(),
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteLaxMode,
			})
			ctx = context.WithValue(ctx, sessionKey, newSession)
			c.SetRequest(req.WithContext(ctx))

			logger.Info("rotated successfully")

			return next(c)
		}
	}
}

func rotateSession(
	ctx context.Context,
	manager models.SessionManager,
	store models.SessionStore,
	oldSession models.Session,
) (models.Session, error) {
	newSession, err := manager.RotateSession(ctx, oldSession)
	if err != nil {
		if errors.Is(err, models.ErrMissingRefreshToken) || errors.Is(err, models.ErrTokenInvalidated) {
			return models.Session{}, common.NewUnauthorizedError("invalid-session", "failed to rotate session").WithInternalError(err)
		}

		return models.Session{},
			common.NewInternalServerError(
				"rotated-failed",
				"failed to rotate session",
			).WithInternalError(err)
	}

	err = store.UpsertSession(ctx, newSession)
	if err != nil {
		return models.Session{}, common.NewInternalServerError("failed-to-store-sesison", "failed to store rotate session").WithInternalError(err)
	}

	return newSession, nil
}

func SessionFromCtx(ctx context.Context) models.Session {
	session, _ := ctx.Value(sessionKey).(models.Session)
	return session
}
