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

	var err error
	treasuryClient, err = client.NewClientWithResponses(testutils.BaseURL)
	if err != nil {
		cancel()
		panic(err)
	}

	code := m.Run()
	cancel()
	os.Exit(code)
}
