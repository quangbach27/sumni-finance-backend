package domain_test

import (
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankFundProviderMetadata(t *testing.T) {
	t.Run("NewBankFundProviderMetadata", func(t *testing.T) {
		tests := []struct {
			name          string
			accountNumber string
			bankName      string
			wantErr       string
		}{
			{
				name:          "reject-empty-account-number",
				accountNumber: "",
				bankName:      "Techcombank",
				wantErr:       "account number can't be empty",
			},
			{
				name:          "reject-empty-bank-name",
				accountNumber: "7777777316",
				bankName:      "",
				wantErr:       "bank name can't be empty",
			},
			{
				name:          "accept-create-bank-metadata",
				accountNumber: "7777777316",
				bankName:      "Techcombank",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				bankMetadata, err := domain.NewBankFundProviderMetadata(tt.accountNumber, tt.bankName)
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, bankMetadata.AccountNumber(), tt.accountNumber)
				assert.Equal(t, bankMetadata.BankName(), tt.bankName)
			})
		}
	})

	t.Run("IsZero", func(t *testing.T) {
		t.Parallel()

		bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")
		assert.False(t, bankMetadata.IsZero())

		zeroMetadata := domain.BankFundProviderMetadata{}
		assert.True(t, zeroMetadata.IsZero())
	})

	t.Run("MatchesType", func(t *testing.T) {
		t.Parallel()

		bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")

		bankType := domain.MustNewFundProviderType("BANK")
		cashType := domain.MustNewFundProviderType("CASH")

		assert.True(t, bankMetadata.MatchesType(bankType))
		assert.False(t, bankMetadata.MatchesType(cashType))
	})
}

func TestCashFundProviderMetadata(t *testing.T) {
	t.Run("NewCashFundProviderMetadata", func(t *testing.T) {
		tests := []struct {
			name      string
			ownerName string
			wantErr   string
		}{
			{
				name:      "reject-empty-owner",
				ownerName: "",
				wantErr:   "owner name can't be empty",
			},
			{
				name:      "create-valid-cash-metadata",
				ownerName: "Bui Quang Bach",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				cashMetadata, err := domain.NewCashFundProviderMetadata(tt.ownerName)
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, cashMetadata.OwnerName(), tt.ownerName)
			})
		}
	})

	t.Run("IsZero", func(t *testing.T) {
		t.Parallel()

		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")
		assert.False(t, cashMetadata.IsZero())

		zeroMetadata := domain.CashFundProviderMetadata{}
		assert.True(t, zeroMetadata.IsZero())
	})

	t.Run("MatchesType", func(t *testing.T) {
		t.Parallel()

		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

		bankType := domain.MustNewFundProviderType("BANK")
		cashType := domain.MustNewFundProviderType("CASH")

		assert.True(t, cashMetadata.MatchesType(cashType))
		assert.False(t, cashMetadata.MatchesType(bankType))
	})
}

func TestNewFundProvider(t *testing.T) {
	t.Run("ValidationsErrors", func(t *testing.T) {
		vnd := shared.MustNewCurrency("VND")

		initBalance := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
		bankMetadata := assertValidBankMetadata(t, "7777777316", "Bui Quang Bach")
		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

		tests := []struct {
			name             string
			fpName           string
			fundProviderType domain.FundProviderType
			initBalance      shared.Money
			metadata         domain.FundProviderMetadata
			wantErrDetails   []common.ErrorDetails
		}{
			{
				name:             "reject-empty-name",
				fpName:           "",
				fundProviderType: domain.MustNewFundProviderType("BANK"),
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
				name:             "reject-empty-fund-provider-type",
				fpName:           "Techcombank-Bach",
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
				fundProviderType: domain.MustNewFundProviderType("BANK"),
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
				fundProviderType: domain.MustNewFundProviderType("BANK"),
				initBalance:      initBalance,
				metadata:         domain.BankFundProviderMetadata{},
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
				fundProviderType: domain.MustNewFundProviderType("BANK"),
				initBalance:      initBalance,
				metadata:         domain.CashFundProviderMetadata{},
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
				fundProviderType: domain.MustNewFundProviderType("BANK"),
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
				fundProviderType: domain.MustNewFundProviderType("CASH"),
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

		t.Run("create bank fund provider sucessfully", func(t *testing.T) {
			t.Parallel()

			bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")

			fp, err := domain.NewFundProvider(
				"Techcombank-Bach",
				domain.MustNewFundProviderType("BANK"),
				initBalance,
				bankMetadata,
			)
			require.NoError(t, err)

			assert.False(t, fp.UUID().IsZero())
			assert.Equal(t, "Techcombank-Bach", fp.Name())
			assert.Equal(t, fp.Type().String(), "BANK")
			assert.True(t, fp.Balance().Equal(initBalance))
			assert.True(t, fp.AvailableMoney().Equal(initBalance))
			assert.True(t, fp.Currency().Equal(initBalance.Currency()))

			assert.Equal(t, fp.Metadata(), bankMetadata)
			assert.Equal(t, fp.Metadata().AccountNumber(), "7777777316")
			assert.Equal(t, fp.Metadata().BankName(), "Techcombank")
		})

		t.Run("create cash fund provider sucessfully", func(t *testing.T) {
			t.Parallel()

			cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

			fp, err := domain.NewFundProvider(
				"Techcombank-Bach",
				domain.MustNewFundProviderType("CASH"),
				initBalance,
				cashMetadata,
			)
			require.NoError(t, err)

			assert.False(t, fp.UUID().IsZero())
			assert.Equal(t, "Techcombank-Bach", fp.Name())
			assert.Equal(t, fp.Type().String(), "CASH")
			assert.True(t, fp.Balance().Equal(initBalance))
			assert.True(t, fp.AvailableMoney().Equal(initBalance))
			assert.True(t, fp.Currency().Equal(initBalance.Currency()))

			assert.Equal(t, fp.Metadata(), cashMetadata)
			assert.Equal(t, fp.Metadata().OwnerName(), "Huynh Trang")
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

func TestFundProvider_TopUp(t *testing.T) {
	bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")

	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")
	initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

	t.Run("reject-when-top-up-empty-money", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		wantErr := fp.TopUp(shared.Money{})

		require.Error(t, wantErr)
		assert.Contains(t, wantErr.Error(), "money for top up can't be empty")
	})

	t.Run("reject-when-top-up-zero-amount", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m1 := assertValidMoney(t, decimal.NewFromInt(0), vnd)

		m2 := assertValidMoney(t, decimal.Zero, vnd)

		wantErr1 := fp.TopUp(m1)
		require.Error(t, wantErr1)

		wantErr2 := fp.TopUp(m2)
		require.Error(t, wantErr2)
	})

	t.Run("reject-when-top-up-negative-amount", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(-100_000), vnd)

		err = fp.TopUp(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "money for top up can't be negative")
	})

	t.Run("reject-when-top-up-money-with-different-currency", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance, // VND
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(100_000), krw) // KRW does not match with Fund Provider currency (VND)

		wantErr := fp.TopUp(m)
		require.Error(t, wantErr)
	})

	t.Run("top-up-money-sucessfully", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)

		err = fp.TopUp(m)
		require.NoError(t, err)

		wantBalance := assertValidMoney(t, decimal.NewFromInt(1_100_000), vnd)

		assert.True(t, fp.Balance().Equal(wantBalance))
	})
}

func TestFundProvider_Withdraw(t *testing.T) {
	bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")

	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")
	initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

	t.Run("reject-when-withdraw-empty-money", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)

		require.NoError(t, err)
		wantErr := fp.Withdraw(shared.Money{})
		require.Error(t, wantErr)
		assert.Contains(t, wantErr.Error(), "money for withdraw can't be empty")
	})

	t.Run("reject-when-withdraw-zero-amount", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.Zero, vnd)

		wantErr := fp.Withdraw(m)
		require.Error(t, wantErr)
	})

	t.Run("reject-when-withdraw-negative-amount", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(-100_000), vnd)

		err = fp.Withdraw(m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "money for withdraw can't be negative")
	})

	t.Run("reject-when-top-up-money-with-different-currency", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(100_000), krw)
		require.NoError(t, err)

		wantErr := fp.TopUp(m)
		require.Error(t, wantErr)
	})

	t.Run("top-up-money-sucess", func(t *testing.T) {
		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		m := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)

		err = fp.Withdraw(m)
		require.NoError(t, err)

		wantBalance := assertValidMoney(t, decimal.NewFromInt(900_000), vnd)

		assert.True(t, fp.Balance().Equal(wantBalance))
	})
}

func TestFundProvider_Reserve(t *testing.T) {
	bankMetadata := assertValidBankMetadata(t, "7777777316", "Techcombank")

	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")
	t.Run("Validations Errors", func(t *testing.T) {
		tests := []struct {
			name           string
			allocatedMoney func(t *testing.T) shared.Money
			wantErr        string
		}{
			{
				name: "reject-allocation-money-currency-mismatch",
				allocatedMoney: func(t *testing.T) shared.Money {
					m := assertValidMoney(t, decimal.NewFromInt(1_000_000), krw)

					return m
				},
				wantErr: "allocation money currency must match fund provider currency",
			},
			{
				name: "reject-allocation-money-exceed-fund-provider-unallocated",
				allocatedMoney: func(t *testing.T) shared.Money {
					m := assertValidMoney(t, decimal.NewFromInt(2_000_000), vnd)

					return m
				},
				wantErr: "allocation money can't exceed fund provider unallocated",
			},
			{
				name: "reject-negative-allocation-money",
				allocatedMoney: func(t *testing.T) shared.Money {
					m := assertValidMoney(t, decimal.NewFromInt(-10_000), vnd)

					return m
				},
				wantErr: "allocation money can't be negative",
			},
			{
				name: "accept-allocation-money-zero",
				allocatedMoney: func(t *testing.T) shared.Money {
					m := assertValidMoney(t, decimal.Zero, vnd)

					return m
				},
			},
			{
				name: "accept-allocation-money",
				allocatedMoney: func(t *testing.T) shared.Money {
					m := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)

					return m
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

				fp, err := domain.NewFundProvider(
					"Techcombank-Bach",
					domain.MustNewFundProviderType("BANK"),
					initBalance,
					bankMetadata,
				)
				require.NoError(t, err)

				allocatedMoney := tt.allocatedMoney(t)
				originalUnallocated := fp.AvailableMoney()

				err = fp.Reserve(allocatedMoney)

				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr)
					return
				}

				require.NoError(t, err)
				wantNewUnallocation, err := originalUnallocated.Sub(allocatedMoney)
				require.NoError(t, err)
				assert.True(t, fp.AvailableMoney().Equal(wantNewUnallocation))
			})
		}
	})

	t.Run("Reserve Success", func(t *testing.T) {
		initBalance, err := shared.NewMoney(decimal.NewFromInt(1_000_000), vnd)
		require.NoError(t, err)

		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		allocationMoney, err := shared.NewMoney(decimal.NewFromInt(600_000), vnd)
		require.NoError(t, err)

		err = fp.Reserve(allocationMoney)
		require.NoError(t, err)

		wantUnallocated, err := shared.NewMoney(decimal.NewFromInt(400_000), vnd)
		require.NoError(t, err)
		assert.True(t, fp.AvailableMoney().Equal(wantUnallocated))
	})

	t.Run("Reserve Zero Amount Success", func(t *testing.T) {
		initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

		fp, err := domain.NewFundProvider(
			"Techcombank-Bach",
			domain.MustNewFundProviderType("BANK"),
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		zeroMoney := assertValidMoney(t, decimal.Zero, vnd)

		err = fp.Reserve(zeroMoney)
		require.NoError(t, err)

		assert.True(t, fp.AvailableMoney().Equal(initBalance))
	})
}

func assertValidBankMetadata(t *testing.T, accountNo string, bankName string) domain.BankFundProviderMetadata {
	bankMetadata, err := domain.NewBankFundProviderMetadata(accountNo, bankName)
	require.NoError(t, err)

	return bankMetadata
}

func assertValidCashMetadata(t *testing.T, ownerName string) domain.CashFundProviderMetadata {
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
