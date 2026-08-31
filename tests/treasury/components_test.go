//go:build component

package treasury_test

import (
	"context"
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
	"sumni-finance-backend/internal/treasury/ports/http/client"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreateFundSource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		req  client.CreateFundSourceRequest
	}{
		{
			name: "cash",
			req: client.CreateFundSourceRequest{
				Name:        "Cash-Wallet",
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.NewFromInt(500_000),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					CashMetadata: &client.FundSourceCashMetadata{
						OwnerName: "Nguyen Van A",
					},
				},
			},
		},
		{
			name: "bank",
			req: client.CreateFundSourceRequest{
				Name:        "Techcombank-SRB",
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.NewFromInt(500_000),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					BankMetadata: &client.FundSourceBankMetadata{
						BankCode: "970436",
						// Randomized so repeated test runs against a shared, non-truncated DB don't
						// collide with the uq_fund_sources_bank_account (bank_account_number, bin) index.
						AccountNumber: gofakeit.Numerify("##########"),
						AccountOwner:  "Nguyen Van A",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fundSourceUUID := createFundSource(ctx, t, treasuryClient, tt.req)
			require.False(t, fundSourceUUID.IsZero())
		})
	}
}

func TestCreateFundSource_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		req          client.CreateFundSourceRequest
		expectedSlug string
	}{
		{
			name: "empty cash owner name",
			req: client.CreateFundSourceRequest{
				Name:        "Cash-Wallet",
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.NewFromInt(500_000),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					CashMetadata: &client.FundSourceCashMetadata{
						OwnerName: "",
					},
				},
			},
			expectedSlug: "invalid-cash-metadata",
		},
		{
			name: "empty bank fields",
			req: client.CreateFundSourceRequest{
				Name:        "Techcombank-SRB",
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.NewFromInt(500_000),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					BankMetadata: &client.FundSourceBankMetadata{
						BankCode:      "",
						AccountNumber: "",
						AccountOwner:  "",
					},
				},
			},
			expectedSlug: "invalid-bank-metadata",
		},
		{
			name: "negative init balance",
			req: client.CreateFundSourceRequest{
				Name:        "Cash-Wallet",
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.NewFromInt(-100),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					CashMetadata: &client.FundSourceCashMetadata{
						OwnerName: "Nguyen Van A",
					},
				},
			},
			expectedSlug: "invalid-fund-source-metadata",
		},
		{
			name: "empty name",
			req: client.CreateFundSourceRequest{
				Name:        "",
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.NewFromInt(500_000),
				Currency:    shared.MustNewCurrency("VND"),
				Metadata: client.FundSourceMetadata{
					CashMetadata: &client.FundSourceCashMetadata{
						OwnerName: "Nguyen Van A",
					},
				},
			},
			expectedSlug: "invalid-fund-source-metadata",
		},
		{
			name: "empty currency",
			req: client.CreateFundSourceRequest{
				Name:        "Cash-Wallet",
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.NewFromInt(500_000),
				Metadata: client.FundSourceMetadata{
					CashMetadata: &client.FundSourceCashMetadata{
						OwnerName: "Nguyen Van A",
					},
				},
			},
			// shared.NewMoney's currency.IsZero() check runs before domain.NewFundSource's own
			// currency check, so this is wrapped with the init-balance slug, not
			// "invalid-fund-source-metadata" (see app/command/create_fund_source.go).
			expectedSlug: "invalid-init-balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := createFundSourceExpectBadRequest(ctx, t, treasuryClient, tt.req)
			require.Equal(t, tt.expectedSlug, errResp.Slug)
		})
	}
}

func newCashFundSourceRequest(name, currency string, initBalance decimal.Decimal, ownerName string) client.CreateFundSourceRequest {
	return client.CreateFundSourceRequest{
		Name:        name,
		SourceType:  domain.FundSourceTypeCash,
		InitBalance: initBalance,
		Currency:    shared.MustNewCurrency(currency),
		Metadata: client.FundSourceMetadata{
			CashMetadata: &client.FundSourceCashMetadata{
				OwnerName: ownerName,
			},
		},
	}
}

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()

	t.Run("no allocations", func(t *testing.T) {
		walletUUID := createWallet(ctx, t, treasuryClient, client.CreateWalletRequest{
			Name:     "Main Wallet",
			Currency: shared.MustNewCurrency("VND"),
		})
		require.False(t, walletUUID.IsZero())
	})

	t.Run("with allocation", func(t *testing.T) {
		fsUUID := createFundSource(ctx, t, treasuryClient, newCashFundSourceRequest("Cash-For-Wallet", "VND", decimal.NewFromInt(1_000_000), "Nguyen Van A"))

		walletUUID := createWallet(ctx, t, treasuryClient, client.CreateWalletRequest{
			Name:     "Main Wallet",
			Currency: shared.MustNewCurrency("VND"),
			Allocations: &[]client.CreateWalletAllocationRequest{
				{FundSourceUuid: fsUUID, AllocatedBalance: decimal.NewFromInt(100_000)},
			},
		})
		require.False(t, walletUUID.IsZero())
	})
}

func TestCreateWallet_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		buildReq     func(t *testing.T) client.CreateWalletRequest
		expectedSlug string
	}{
		{
			name: "empty name",
			buildReq: func(t *testing.T) client.CreateWalletRequest {
				return client.CreateWalletRequest{
					Name:     "",
					Currency: shared.MustNewCurrency("VND"),
				}
			},
			expectedSlug: "invalid-wallet",
		},
		{
			name: "empty currency",
			buildReq: func(t *testing.T) client.CreateWalletRequest {
				return client.CreateWalletRequest{
					Name: "My Wallet",
				}
			},
			expectedSlug: "invalid-wallet",
		},
		{
			name: "allocation currency mismatch",
			buildReq: func(t *testing.T) client.CreateWalletRequest {
				fsUUID := createFundSource(ctx, t, treasuryClient, newCashFundSourceRequest("KRW-Cash", "KRW", decimal.NewFromInt(500_000), "Nguyen Van A"))
				return client.CreateWalletRequest{
					Name:     "VND Wallet",
					Currency: shared.MustNewCurrency("VND"),
					Allocations: &[]client.CreateWalletAllocationRequest{
						{FundSourceUuid: fsUUID, AllocatedBalance: decimal.NewFromInt(100_000)},
					},
				}
			},
			expectedSlug: "failed-to-allocate",
		},
		{
			name: "allocation exceeds available balance",
			buildReq: func(t *testing.T) client.CreateWalletRequest {
				fsUUID := createFundSource(ctx, t, treasuryClient, newCashFundSourceRequest("Small-Cash", "VND", decimal.NewFromInt(10_000), "Nguyen Van A"))
				return client.CreateWalletRequest{
					Name:     "VND Wallet",
					Currency: shared.MustNewCurrency("VND"),
					Allocations: &[]client.CreateWalletAllocationRequest{
						{FundSourceUuid: fsUUID, AllocatedBalance: decimal.NewFromInt(999_999)},
					},
				}
			},
			expectedSlug: "failed-to-allocate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := createWalletExpectBadRequest(ctx, t, treasuryClient, tt.buildReq(t))
			require.Equal(t, tt.expectedSlug, errResp.Slug)
		})
	}
}

func TestLinkFundSources(t *testing.T) {
	ctx := context.Background()

	walletUUID := createWallet(ctx, t, treasuryClient, client.CreateWalletRequest{
		Name:     "Wallet-To-Link",
		Currency: shared.MustNewCurrency("VND"),
	})
	fsUUID := createFundSource(ctx, t, treasuryClient, newCashFundSourceRequest("Cash-To-Link", "VND", decimal.NewFromInt(500_000), "Nguyen Van A"))

	linkFundSources(ctx, t, treasuryClient, walletUUID, client.LinkFundSourcesRequest{
		Allocations: []client.LinkFundSourceAllocationRequest{
			{FundSourceUuid: fsUUID, AllocatedAmount: decimal.NewFromInt(50_000)},
		},
	})

	resp := listWallets(ctx, t, treasuryClient, &client.ListWalletsParams{Page: common.ToPtr(1), PageSize: common.ToPtr(200)})
	var found *client.WalletResponse
	for i := range resp.Items {
		if resp.Items[i].WalletUuid == walletUUID {
			found = &resp.Items[i]
			break
		}
	}
	require.NotNil(t, found, "created wallet not found in ListWallets response")
	require.Len(t, found.Allocations, 1)
	require.Equal(t, fsUUID, found.Allocations[0].FundSourceUuid)
	require.True(t, decimal.NewFromInt(50_000).Equal(found.Allocations[0].Balance))
}

func TestLinkFundSources_Validation(t *testing.T) {
	ctx := context.Background()

	walletUUID := createWallet(ctx, t, treasuryClient, client.CreateWalletRequest{
		Name:     "Wallet-For-Validation",
		Currency: shared.MustNewCurrency("VND"),
	})

	errResp := linkFundSourcesExpectBadRequest(ctx, t, treasuryClient, walletUUID, client.LinkFundSourcesRequest{
		Allocations: []client.LinkFundSourceAllocationRequest{},
	})
	require.Equal(t, "invalid-link-fund-sources-command", errResp.Slug)
}

func TestListFundSources(t *testing.T) {
	ctx := context.Background()

	cashUUID := createFundSource(ctx, t, treasuryClient, newCashFundSourceRequest("List-Cash", "VND", decimal.NewFromInt(500_000), "Nguyen Van A"))
	bankUUID := createFundSource(ctx, t, treasuryClient, client.CreateFundSourceRequest{
		Name:        "List-Bank",
		SourceType:  domain.FundSourceTypeBank,
		InitBalance: decimal.NewFromInt(500_000),
		Currency:    shared.MustNewCurrency("VND"),
		Metadata: client.FundSourceMetadata{
			BankMetadata: &client.FundSourceBankMetadata{
				BankCode:      "970436",
				AccountNumber: gofakeit.Numerify("##########"),
				AccountOwner:  "Nguyen Van A",
			},
		},
	})

	t.Run("contains created items", func(t *testing.T) {
		resp := listFundSources(ctx, t, treasuryClient, &client.ListFundSourcesParams{Page: common.ToPtr(1), PageSize: common.ToPtr(200)})
		require.Equal(t, 1, resp.Pagination.Page)
		require.Equal(t, 200, resp.Pagination.PageSize)

		uuids := make(map[client.FundSourceUUID]bool, len(resp.Items))
		for _, item := range resp.Items {
			uuids[item.FundSourceUuid] = true
		}
		require.True(t, uuids[cashUUID], "created cash fund source not found in list")
		require.True(t, uuids[bankUUID], "created bank fund source not found in list")
	})

	t.Run("default pagination", func(t *testing.T) {
		resp := listFundSources(ctx, t, treasuryClient, nil)
		require.Equal(t, 1, resp.Pagination.Page)
		require.Equal(t, 20, resp.Pagination.PageSize)
	})
}

func TestListWallets(t *testing.T) {
	ctx := context.Background()

	walletUUID := createWallet(ctx, t, treasuryClient, client.CreateWalletRequest{
		Name:     "List-Wallet",
		Currency: shared.MustNewCurrency("VND"),
	})

	t.Run("contains created wallet", func(t *testing.T) {
		resp := listWallets(ctx, t, treasuryClient, &client.ListWalletsParams{Page: common.ToPtr(1), PageSize: common.ToPtr(200)})
		require.Equal(t, 1, resp.Pagination.Page)
		require.Equal(t, 200, resp.Pagination.PageSize)

		var found *client.WalletResponse
		for i := range resp.Items {
			if resp.Items[i].WalletUuid == walletUUID {
				found = &resp.Items[i]
				break
			}
		}
		require.NotNil(t, found, "created wallet not found in ListWallets response")
		require.Equal(t, "List-Wallet", found.Name)
	})

	t.Run("default pagination", func(t *testing.T) {
		resp := listWallets(ctx, t, treasuryClient, nil)
		require.Equal(t, 1, resp.Pagination.Page)
		require.Equal(t, 20, resp.Pagination.PageSize)
	})
}
