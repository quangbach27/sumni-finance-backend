package domain_test

import (
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/finance/app/models"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFundProvider(t *testing.T) {
	t.Run("ValidationsErrors", func(t *testing.T) {
		vnd := shared.MustNewCurrency("VND")

		validOfficeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}
		initBalance := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
		bankMetadata := validBankMetadata(t, "7777777316", "Bui Quang Bach")
		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

		tests := []struct {
			name             string
			fpName           string
			officeUUID       models.OfficeUUID
			fundProviderType domain.FundProviderType
			initBalance      shared.Money
			metadata         domain.FundProviderMetadata
			wantErrDetails   []common.ErrorDetails
		}{
			{
				name:             "reject-empty-name",
				fpName:           "",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      initBalance,
				metadata:         bankMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-name",
						Message:    "name can't be empty",
					},
				},
			},
			{
				name:             "reject-empty-office-uuid",
				fpName:           "Techcombank-Bach",
				officeUUID:       models.OfficeUUID{},
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      initBalance,
				metadata:         bankMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-office-uuid",
						Message:    "office uuid can't be empty",
					},
				},
			},
			{
				name:             "reject-empty-fund-provider-type",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderType{},
				initBalance:      initBalance,
				metadata:         bankMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-fund-provider-type",
						Message:    "fund provider type can't be empty",
					},
					{
						EntityType: "fund_provider",
						ErrorSlug:  "metadata-mismatch",
					},
				},
			},
			{
				name:             "reject-empty-init-balance",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      shared.Money{},
				metadata:         bankMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-balance",
						Message:    "balance can't be empty",
					},
				},
			},
			{
				name:             "reject-empty-bank-fund-provider-metadata",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      initBalance,
				metadata:         domain.FundProviderBankMetadata{},
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-metadata",
						Message:    "metadata can't not be empty",
					},
				},
			},
			{
				name:             "reject-empty-cash-fund-provider-metadata",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      initBalance,
				metadata:         domain.FundProviderCashMetadata{},
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "empty-metadata",
						Message:    "metadata can't not be empty",
					},
				},
			},
			{
				name:             "reject-cash-metadata-for-non-cash-provider",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeBank,
				initBalance:      initBalance,
				metadata:         cashMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "metadata-mismatch",
					},
				},
			},
			{
				name:             "reject-bank-metadata-for--non-bank-provider",
				fpName:           "Techcombank-Bach",
				officeUUID:       validOfficeUUID,
				fundProviderType: domain.FundProviderTypeCash,
				initBalance:      initBalance,
				metadata:         bankMetadata,
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "fund_provider",
						ErrorSlug:  "metadata-mismatch",
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := domain.NewFundProvider(
					tt.fpName,
					tt.officeUUID,
					tt.fundProviderType,
					tt.initBalance,
					tt.metadata,
				)
				require.Error(t, err)
				testutils.AssertErrorDetails(t, err, tt.wantErrDetails)
			})
		}
	})

	t.Run("Sucess", func(t *testing.T) {
		vnd := shared.MustNewCurrency("VND")
		initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
		officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

		t.Run("create bank fund provider sucessfully", func(t *testing.T) {
			t.Parallel()

			bankMetadata := validBankMetadata(t, "7777777316", "Bui Quang Bach")

			fp, err := domain.NewFundProvider(
				"Techcombank-Bach",
				officeUUID,
				domain.FundProviderTypeBank,
				initBalance,
				bankMetadata,
			)
			require.NoError(t, err)

			assert.False(t, fp.UUID().IsZero())
			assert.Equal(t, officeUUID, fp.OfficeUUID())
			assert.Equal(t, "Techcombank-Bach", fp.Name())
			assert.Equal(t, fp.Type(), domain.FundProviderTypeBank)
			assert.True(t, fp.Balance().Equal(initBalance))
			assert.True(t, fp.AvailableBalance().Equal(initBalance))
			assert.True(t, fp.Currency().Equal(initBalance.Currency()))

			bankMeta, ok := fp.BankMetadata()
			require.True(t, ok)
			assert.Equal(t, bankMeta, bankMetadata)
			assert.Equal(t, bankMeta.AccountNumber(), "7777777316")
			assert.Equal(t, bankMeta.AccountOwner(), "Bui Quang Bach")
		})

		t.Run("create cash fund provider sucessfully", func(t *testing.T) {
			t.Parallel()

			cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

			fp, err := domain.NewFundProvider(
				"Techcombank-Bach",
				officeUUID,
				domain.FundProviderTypeCash,
				initBalance,
				cashMetadata,
			)
			require.NoError(t, err)

			assert.False(t, fp.UUID().IsZero())
			assert.Equal(t, officeUUID, fp.OfficeUUID())
			assert.Equal(t, "Techcombank-Bach", fp.Name())
			assert.Equal(t, fp.Type(), domain.FundProviderTypeCash)
			assert.True(t, fp.Balance().Equal(initBalance))
			assert.True(t, fp.AvailableBalance().Equal(initBalance))
			assert.True(t, fp.Currency().Equal(initBalance.Currency()))

			cashMeta, ok := fp.CashMetadata()
			require.True(t, ok)
			assert.Equal(t, cashMeta, cashMetadata)
			assert.Equal(t, cashMeta.OwnerName(), "Huynh Trang")
		})
	})

	t.Run("Withdraw", func(t *testing.T) {
		tests := []struct {
			name string
		}{}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
			})
		}
	})

	t.Run("Allocate", func(t *testing.T) {
		tests := []struct {
			name string
		}{}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
			})
		}
	})
}

func validBankMetadata(t *testing.T, accountNo string, accountOwner string) domain.FundProviderBankMetadata {
	t.Helper()
	bankMetadata, err := domain.NewFundProviderBankMetadata(newBankInfoForMetadataTest(t), accountNo, accountOwner)
	require.NoError(t, err)
	return bankMetadata
}

func assertValidCashMetadata(t *testing.T, ownerName string) domain.FundProviderCashMetadata {
	cashMetadata, err := domain.NewCashFundProviderMetadata(ownerName)
	require.NoError(t, err)

	return cashMetadata
}

func assertValidMoney(t *testing.T, amount decimal.Decimal, currency shared.Currency) shared.Money {
	t.Helper()
	money, err := shared.NewMoney(amount, currency)
	require.NoError(t, err)
	return money
}
