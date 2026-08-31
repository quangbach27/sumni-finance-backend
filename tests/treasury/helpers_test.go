//go:build component

package treasury_test

import (
	"context"
	"net/http"
	"testing"

	"sumni-finance-backend/internal/treasury/ports/http/client"

	"github.com/stretchr/testify/require"
)

func createFundSource(ctx context.Context, t *testing.T, c *client.ClientWithResponses, req client.CreateFundSourceRequest) client.FundSourceUUID {
	t.Helper()

	resp, err := c.CreateFundSourceWithResponse(ctx, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.FundSourceUuid
}

func createFundSourceExpectBadRequest(ctx context.Context, t *testing.T, c *client.ClientWithResponses, req client.CreateFundSourceRequest) *client.ErrorResponse {
	t.Helper()

	resp, err := c.CreateFundSourceWithResponse(ctx, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON400)

	return resp.JSON400
}

func createWallet(ctx context.Context, t *testing.T, c *client.ClientWithResponses, req client.CreateWalletRequest) client.WalletUUID {
	t.Helper()

	resp, err := c.CreateWalletWithResponse(ctx, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.WalletUuid
}

func createWalletExpectBadRequest(ctx context.Context, t *testing.T, c *client.ClientWithResponses, req client.CreateWalletRequest) *client.ErrorResponse {
	t.Helper()

	resp, err := c.CreateWalletWithResponse(ctx, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON400)

	return resp.JSON400
}

func linkFundSources(ctx context.Context, t *testing.T, c *client.ClientWithResponses, walletUUID client.WalletUUID, req client.LinkFundSourcesRequest) {
	t.Helper()

	resp, err := c.LinkFundSourcesWithResponse(ctx, walletUUID, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), string(resp.Body))
}

func linkFundSourcesExpectBadRequest(ctx context.Context, t *testing.T, c *client.ClientWithResponses, walletUUID client.WalletUUID, req client.LinkFundSourcesRequest) *client.ErrorResponse {
	t.Helper()

	resp, err := c.LinkFundSourcesWithResponse(ctx, walletUUID, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON400)

	return resp.JSON400
}

func listFundSources(ctx context.Context, t *testing.T, c *client.ClientWithResponses, params *client.ListFundSourcesParams) *client.ListFundSourcesResponse {
	t.Helper()

	resp, err := c.ListFundSourcesWithResponse(ctx, params)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON200)

	return resp.JSON200
}

func listWallets(ctx context.Context, t *testing.T, c *client.ClientWithResponses, params *client.ListWalletsParams) *client.ListWalletsResponse {
	t.Helper()

	resp, err := c.ListWalletsWithResponse(ctx, params)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON200)

	return resp.JSON200
}
