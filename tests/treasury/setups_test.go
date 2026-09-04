//go:build component

package treasury_test

import (
	"context"
	"os"
	"testing"

	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury/ports/http/client"
)

var treasuryClient *client.ClientWithResponses

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())

	testutils.StartServer(ctx)

	clients, err := testutils.NewTestClients()
	if err != nil {
		cancel()
		panic(err)
	}
	treasuryClient = clients.Treasury

	code := m.Run()
	cancel()
	testutils.CloseDB()
	os.Exit(code)
}
