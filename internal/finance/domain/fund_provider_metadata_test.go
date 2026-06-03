package domain_test

import (
	"testing"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFundProviderBankMetadata(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validBankInfo := newBankInfoForMetadataTest(t)

		tests := []struct {
			name          string
			bankInfo      shared.BankInfo
			accountNumber string
			accountOwner  string
			wantErr       string
		}{
			{
				name:          "reject-zero-bank-info",
				bankInfo:      shared.BankInfo{},
				accountNumber: "7777777316",
				accountOwner:  "Bui Quang Bach",
				wantErr:       "bank info can't be empty",
			},
			{
				name:          "reject-empty-account-number",
				bankInfo:      validBankInfo,
				accountNumber: "",
				accountOwner:  "Bui Quang Bach",
				wantErr:       "bank info can't be empty",
			},
			{
				name:          "reject-empty-account-owner",
				bankInfo:      validBankInfo,
				accountNumber: "7777777316",
				accountOwner:  "",
				wantErr:       "account owner can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewFundProviderBankMetadata(tt.bankInfo, tt.accountNumber, tt.accountOwner)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		bankInfo := newBankInfoForMetadataTest(t)

		metadata, err := domain.NewFundProviderBankMetadata(bankInfo, "7777777316", "Bui Quang Bach")
		require.NoError(t, err)

		assert.Equal(t, bankInfo, metadata.BankInfo())
		assert.Equal(t, "7777777316", metadata.AccountNumber())
		assert.Equal(t, "Bui Quang Bach", metadata.AccountOwner())
	})

	t.Run("IsZero", func(t *testing.T) {
		t.Parallel()

		bankInfo := newBankInfoForMetadataTest(t)
		metadata, err := domain.NewFundProviderBankMetadata(bankInfo, "7777777316", "Bui Quang Bach")
		require.NoError(t, err)

		assert.False(t, metadata.IsZero())
		assert.True(t, domain.FundProviderBankMetadata{}.IsZero())
	})

	t.Run("MatchesType", func(t *testing.T) {
		t.Parallel()

		bankInfo := newBankInfoForMetadataTest(t)
		metadata, err := domain.NewFundProviderBankMetadata(bankInfo, "7777777316", "Bui Quang Bach")
		require.NoError(t, err)

		assert.True(t, metadata.MatchesType(domain.FundProviderTypeBank))
		assert.False(t, metadata.MatchesType(domain.FundProviderTypeCash))
	})
}

func TestNewCashFundProviderMetadata(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		tests := []struct {
			name      string
			ownerName string
			wantErr   string
		}{
			{
				name:      "reject-empty-owner-name",
				ownerName: "",
				wantErr:   "owner name can't be empty",
			},
			{
				name:      "reject-whitespace-only-owner-name",
				ownerName: "   ",
				wantErr:   "owner name can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewCashFundProviderMetadata(tt.ownerName)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		metadata, err := domain.NewCashFundProviderMetadata("Bui Quang Bach")
		require.NoError(t, err)

		assert.Equal(t, "Bui Quang Bach", metadata.OwnerName())
	})

	t.Run("trims-whitespace-from-owner-name", func(t *testing.T) {
		t.Parallel()

		metadata, err := domain.NewCashFundProviderMetadata("  Bui Quang Bach  ")
		require.NoError(t, err)

		assert.Equal(t, "Bui Quang Bach", metadata.OwnerName())
	})

	t.Run("IsZero", func(t *testing.T) {
		t.Parallel()

		metadata, err := domain.NewCashFundProviderMetadata("Bui Quang Bach")
		require.NoError(t, err)

		assert.False(t, metadata.IsZero())
		assert.True(t, domain.FundProviderCashMetadata{}.IsZero())
	})

	t.Run("MatchesType", func(t *testing.T) {
		t.Parallel()

		metadata, err := domain.NewCashFundProviderMetadata("Bui Quang Bach")
		require.NoError(t, err)

		assert.True(t, metadata.MatchesType(domain.FundProviderTypeCash))
		assert.False(t, metadata.MatchesType(domain.FundProviderTypeBank))
	})
}

func newBankInfoForMetadataTest(t *testing.T) shared.BankInfo {
	t.Helper()
	bankInfo, err := shared.NewBankInfo("Vietcombank", 970436, "VCB", "VCB", "https://logo.url", "https://icon.url", 1, 1)
	require.NoError(t, err)
	return bankInfo
}
