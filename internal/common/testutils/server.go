package testutils

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"sumni-finance-backend/internal"
	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/treasury/ports/http/client"
)

const BaseURL = "http://localhost:9090"

type TestClients struct {
	Treasury *client.ClientWithResponses
}

func NewTestClients() (TestClients, error) {
	treasuryClient, err := client.NewClientWithResponses(BaseURL)
	if err != nil {
		return TestClients{}, err
	}

	return TestClients{Treasury: treasuryClient}, nil
}

func StartServer(ctx context.Context) {
	log.Init(slog.LevelInfo)

	config := common.NewConfig()

	svc, err := internal.New(
		ctx,
		config,
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
		resp, err := http.Get(BaseURL + "/healthz")
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
