package treasury_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/testutils"
	treasuryClient "sumni-finance-backend/internal/treasury/api/http/client"
)

type testClient = *treasuryClient.ClientWithResponses

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testutils.StartServer(ctx)

	os.Exit(m.Run())
}

func newTestClient(t *testing.T) testClient {
	t.Helper()

	return testutils.NewTestClients(t).Treasury
}

func assertCorrelationID(t *testing.T, resp *http.Response) {
	t.Helper()
	require.NotNil(t, resp, "HTTP response must not be nil")
	assert.NotEmpty(t, resp.Header.Get("Correlation-ID"), "response must include Correlation-ID header")
}
