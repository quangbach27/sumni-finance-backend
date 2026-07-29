package testutils

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"sumni-finance-backend/internal"
	"sumni-finance-backend/internal/common/log"
)

const BaseURL = "http://localhost:9090"

type TestClients struct{}

func NewTestClients(t *testing.T) TestClients {
	t.Helper()

	return TestClients{}
}

func StartServer(ctx context.Context) {
	log.Init(slog.LevelInfo)

	svc, err := internal.New(
		ctx,
		NewDB(),
		internal.ExternalService{},
	)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := svc.Run(ctx, ":9090"); err != nil {
			panic(err)
		}
	}()

	waitForServer()
}

func waitForServer() {
	for range 100 {
		resp, err := http.Get(BaseURL + "/health")
		if err == nil && resp.StatusCode < 300 {
			_ = resp.Body.Close()
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	panic("server did not start within 5 seconds")
}
