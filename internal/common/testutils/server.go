package testutils

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"sumni-finance-backend/internal"
	commonHTTP "sumni-finance-backend/internal/common/http"
	"sumni-finance-backend/internal/common/log"
	bankLookup "sumni-finance-backend/internal/treasury/adapters/bank/lookup"
	treasuryClient "sumni-finance-backend/internal/treasury/api/http/client"

	"github.com/stretchr/testify/require"
)

const BaseURL = "http://localhost:9090"

type TestClients struct {
	Treasury *treasuryClient.ClientWithResponses
}

func NewTestClients(t *testing.T) TestClients {
	t.Helper()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	editorFn := func(ctx context.Context, req *http.Request) error {
		log.FromContext(ctx).
			With(
				"method", req.Method,
				"url", req.URL.String(),
				"test_name", t.Name(),
			).
			Info("Making component test API request")

		req.Header.Set(commonHTTP.TestNameHeader, t.Name())
		return nil
	}

	treasury, err := treasuryClient.NewClientWithResponses(
		BaseURL,
		treasuryClient.WithHTTPClient(httpClient),
		treasuryClient.WithRequestEditorFn(editorFn),
	)
	require.NoError(t, err, "creating treasury client failed")

	return TestClients{
		Treasury: treasury,
	}
}

func StartServer(ctx context.Context) {
	log.Init(slog.LevelInfo)

	db, _ := NewDB(ctx)
	svc, err := internal.New(
		ctx,
		db,
		internal.ExternalService{
			BankLookupProvider: bankLookup.NewStubClient(),
		},
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
