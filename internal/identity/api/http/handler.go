package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/identity/app/models"

	"github.com/labstack/echo/v4"
)

type AuthRequest struct {
	Url      string
	State    string
	Verifier string
}

type CallbackInput struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
	ExpectedState    string
	Verifier         string
}

type Authenticator interface {
	LoginURL() (AuthRequest, error)
	HandleCallback(ctx context.Context, input CallbackInput) (models.Session, error)
	LogoutURL(idTokenHint, postLogoutRedirectURI string) (string, error)
}

const (
	flowCookieName    = "kc_flow"
	flowCookieTTl     = 3 * time.Minute
	sessionCookieName = "session_id"
)

type Handlers struct {
	config       *common.Config
	auth         Authenticator
	sessionStore models.SessionStore
}

func NewHandlers(
	config *common.Config,
	authenticator Authenticator,
	sessionStore models.SessionStore,
) Handlers {
	if config == nil {
		panic("config can't be nil")
	}

	if authenticator == nil {
		panic("authenticator can't be nil")
	}

	if sessionStore == nil {
		panic("session store can't be nil")
	}

	return Handlers{
		config:       config,
		auth:         authenticator,
		sessionStore: sessionStore,
	}
}

func (h Handlers) Login(c echo.Context) error {
	req, err := h.auth.LoginURL()
	slog.Info(req.Url)
	if err != nil {
		return common.Error{
			HttpErrorCode: http.StatusInternalServerError,
			PublicError:   "failed to start login",
			ErrorSlug:     "failed-to-start-login",
			InternalError: err,
		}
	}

	c.SetCookie(&http.Cookie{
		Name:     flowCookieName,
		Value:    req.State + "." + req.Verifier,
		Path:     "/",
		Expires:  time.Now().Add(flowCookieTTl),
		HttpOnly: true,
		Secure:   false, // requires HTTPS; set false only for local http dev
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusFound, req.Url)
}

func (h Handlers) Callback(c echo.Context) error {
	ctx := c.Request().Context()

	state, verifier, err := h.readAndClearFlowCookie(c)
	if err != nil {
		return err
	}

	session, err := h.auth.HandleCallback(
		ctx,
		CallbackInput{
			Code:             c.QueryParam("code"),
			State:            c.QueryParam("state"),
			Error:            c.QueryParam("error"),
			ErrorDescription: c.QueryParam("error_description"),
			ExpectedState:    state,
			Verifier:         verifier,
		},
	)
	if err != nil {
		return common.NewUnauthorizedError(
			"failed-to-handle-callback",
			"failed to handle authorization callback",
		).WithInternalError(err)
	}

	err = h.sessionStore.UpsertSession(ctx, session)
	if err != nil {
		return common.
			NewInternalServerError(
				"failed-to-save-session-id",
				"failed to save session id",
			).
			WithInternalError(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	log.FromContext(ctx).Info("user authenticated successfully")
	return c.Redirect(http.StatusFound, h.config.Auth.PostLoginRedirect)
}

func (h Handlers) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	sessionID, err := getSessionIDFromCookie(c)
	if err != nil {
		return err
	}

	session, err := h.sessionStore.GetActiveSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	err = h.sessionStore.ExpireSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	clearCookie(c, sessionCookieName)

	logoutUrl, err := h.auth.LogoutURL(session.IDToken, h.config.Auth.PostLogoutRedirect)
	if err != nil {
		return common.NewInternalServerError("failed-to-logout", "failed to logout").WithInternalError(err)
	}

	log.FromContext(ctx).Info("user logged out successfully")
	return c.Redirect(http.StatusFound, logoutUrl)
}

func (h Handlers) readAndClearFlowCookie(c echo.Context) (state, verifier string, err error) {
	cookie, err := c.Cookie(flowCookieName)
	if err != nil {
		return "", "", common.NewInvalidInputError("missing-cookie", "missing or expired flow cookie").WithInternalError(err)
	}

	clearCookie(c, flowCookieName)

	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", common.NewInvalidInputError("malformed", "malformed flow cookie")
	}

	return parts[0], parts[1], nil
}

func getSessionIDFromCookie(c echo.Context) (models.SessionID, error) {
	sessionCookie, err := c.Cookie(sessionCookieName)
	if err != nil {
		return models.SessionID{}, common.NewUnauthorizedError(
			"unauthorized",
			"can not retrieve session id",
		).WithInternalError(err)
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		return models.SessionID{}, common.NewUnauthorizedError(
			"unauthorized",
			"empty session id",
		).WithInternalError(err)
	}

	return models.SessionID{
		UUID: common.MustUUIDFromString(sessionCookie.Value),
	}, nil
}

func clearCookie(c echo.Context, cookieName string) {
	c.SetCookie(&http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
}

func Register(
	config *common.Config,
	publicRoute common.EchoRouter,
	protectedRoute common.EchoRouter,
	handlers Handlers,
) error {
	publicRoute.GET("/auth/login", handlers.Login)
	publicRoute.GET("/auth/callback", handlers.Callback)
	protectedRoute.GET("/auth/logout", handlers.Logout)
	return nil
}
