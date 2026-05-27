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

func TestNewWallet(t *testing.T) {
	bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")

	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")

	officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}
	initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			testName        string
			wName           string
			description     string
			currency        shared.Currency
			officeUUID      models.OfficeUUID
			allocationsData func(
				t *testing.T,
				fp1 *domain.FundProvider,
			) []domain.NewFundProviderAllocationData
			wantErrDetails []common.ErrorDetails
		}{
			{
				testName:    "reject-empty-name",
				wName:       "",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					allocated := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "wallet",
						ErrorSlug:  "empty-name",
						Message:    "name can't be empty",
					},
				},
			},
			{
				testName:    "reject-empty-currency",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    shared.Currency{},
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					allocated := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "wallet",
						ErrorSlug:  "empty-currency",
						Message:    "currency can't be empty",
					},
				},
			},
			{
				testName:    "reject-empty-office-uuid",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  models.OfficeUUID{},
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					return []domain.NewFundProviderAllocationData{}
				},
				wantErrDetails: []common.ErrorDetails{
					{
						EntityType: "wallet",
						ErrorSlug:  "empty-office-uuid",
						Message:    "office uuid can't be empty",
					},
				},
			},
			{
				testName:    "reject-nil-fund-provider-in-allocation",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     nil,
							AllocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
						},
					}
				},
			},
			{
				testName:    "reject-office-uuid-missmatch-between-fund-provider-and-wallet",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    krw,
				officeUUID:  models.OfficeUUID{UUID: common.NewUUIDv7()},
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					allocated := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
			},
			{
				testName:    "reject-currency-missmatch-between-fund-provider-and-wallet",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    krw,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					allocated := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
			},
			{
				testName:    "reject-negative-allocation-amount",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()

					allocated := assertValidMoney(t, decimal.NewFromInt(-100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
			},
			{
				testName:    "reject-empty-amount-in-allocation",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: shared.Money{},
						},
					}
				},
			},
			{
				testName:    "reject-duplicate-fund-providers",
				wName:       "Vi tai chinh tong",
				description: "wallet contains church money",
				currency:    vnd,
				officeUUID:  officeUUID,
				allocationsData: func(t *testing.T, fp1 *domain.FundProvider) []domain.NewFundProviderAllocationData {
					t.Helper()
					allocated := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)
					return []domain.NewFundProviderAllocationData{
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
						{
							FundProvider:     fp1,
							AllocationAmount: allocated,
						},
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				fp1, err := domain.NewFundProvider(
					"Techcombank-Bach",
					officeUUID,
					domain.FundProviderTypeBank,
					initBalance,
					bankMetadata,
				)
				require.NoError(t, err)

				require.NoError(t, err)

				_, err = domain.NewWallet(
					tt.wName,
					tt.description,
					tt.currency,
					tt.officeUUID,
					tt.allocationsData(t, fp1),
				)

				require.Error(t, err)
				if len(tt.wantErrDetails) != 0 {
					testutils.AssertErrorDetails(t, err, tt.wantErrDetails)
				}
			})
		}
	})

	t.Run("create-wallet-success", func(t *testing.T) {
		bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")
		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

		vnd := shared.MustNewCurrency("VND")

		initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

		fp1, err := domain.NewFundProvider(
			"Techcombank-Bach",
			officeUUID,
			domain.FundProviderTypeBank,
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		fp2, err := domain.NewFundProvider(
			"Techcombank-Bach",
			officeUUID,
			domain.FundProviderTypeCash,
			initBalance,
			cashMetadata,
		)
		require.NoError(t, err)

		allocationAmount1, _ := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
		allocationAmount2, _ := shared.NewMoney(decimal.NewFromInt(400_000), vnd)

		w, err := domain.NewWallet(
			"Vi Tai Chinh Tong",
			"Wallet contains office money",
			vnd,
			officeUUID,
			[]domain.NewFundProviderAllocationData{
				{
					FundProvider:     fp1,
					AllocationAmount: allocationAmount1,
				},
				{
					FundProvider:     fp2,
					AllocationAmount: allocationAmount2,
				},
			},
		)
		require.NoError(t, err)

		assert.NotNil(t, w)
		assert.Equal(t, w.Name(), "Vi Tai Chinh Tong")
		assert.True(t, w.Currency().Equal(vnd))
		assert.Equal(t, w.Description(), "Wallet contains office money")
		assert.Equal(t, w.OfficeUUID(), officeUUID)

		assert.True(t, w.Balance().Equal(assertValidMoney(t, decimal.NewFromInt(900_000), vnd))) // sum of all allocations in fund provider registry

		assert.True(t, fp1.AvailableBalance().Equal(assertValidMoney(t, decimal.NewFromInt(500_000), vnd)))
		assert.True(t, fp2.AvailableBalance().Equal(assertValidMoney(t, decimal.NewFromInt(600_000), vnd)))
	})

	t.Run("create-wallet-success-with-zero-amount-allocation", func(t *testing.T) {
		bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")
		cashMetadata := assertValidCashMetadata(t, "Huynh Trang")

		vnd := shared.MustNewCurrency("VND")

		initBalance := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)

		fp1, err := domain.NewFundProvider(
			"Techcombank-Bach",
			officeUUID,
			domain.FundProviderTypeBank,
			initBalance,
			bankMetadata,
		)
		require.NoError(t, err)

		fp2, err := domain.NewFundProvider(
			"Techcombank-Bach",
			officeUUID,
			domain.FundProviderTypeCash,
			initBalance,
			cashMetadata,
		)
		require.NoError(t, err)

		allocationAmount1, _ := shared.NewMoney(decimal.Zero, vnd)
		allocationAmount2, _ := shared.NewMoney(decimal.Zero, vnd)

		w, err := domain.NewWallet(
			"Vi Tai Chinh Tong",
			"Wallet contains office money",
			vnd,
			officeUUID,
			[]domain.NewFundProviderAllocationData{
				{
					FundProvider:     fp1,
					AllocationAmount: allocationAmount1,
				},
				{
					FundProvider:     fp2,
					AllocationAmount: allocationAmount2,
				},
			},
		)
		require.NoError(t, err)

		assert.NotNil(t, w)
		assert.Equal(t, w.Name(), "Vi Tai Chinh Tong")
		assert.True(t, w.Currency().Equal(vnd))
		assert.Equal(t, w.Description(), "Wallet contains office money")
		assert.True(t, w.Balance().Equal(assertValidMoney(t, decimal.Zero, vnd)))

		assert.True(t, fp1.AvailableBalance().Equal(initBalance))
		assert.True(t, fp2.AvailableBalance().Equal(initBalance))
	})

	t.Run("create-wallet-success-without-fund-allocation", func(t *testing.T) {
		vnd := shared.MustNewCurrency("VND")

		w, err := domain.NewWallet(
			"Vi Tai Chinh Tong",
			"Wallet contains office money",
			vnd,
			officeUUID,
			nil,
		)
		require.NoError(t, err)

		assert.NotNil(t, w)
		assert.Equal(t, w.Name(), "Vi Tai Chinh Tong")
		assert.True(t, w.Currency().Equal(vnd))
		assert.Equal(t, w.Description(), "Wallet contains office money")
		assert.True(t, w.Balance().Equal(assertValidMoney(t, decimal.Zero, vnd)))
	})
}

func TestWallet_AllocateFundProvider(t *testing.T) {
	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")

	initBalanceVnd := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
	initBalanceKrw := assertValidMoney(t, decimal.NewFromInt(1_000_000), krw)

	officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name                    string
			allocationData          domain.NewFundProviderAllocationData
			officeUUID              models.OfficeUUID
			fundProvider            *domain.FundProvider
			isFundProviderAllocated bool
			allocationAmount        shared.Money
			wantErr                 string
		}{
			{
				name:             "reject-nil-fund-provider",
				fundProvider:     nil,
				officeUUID:       officeUUID,
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
				wantErr:          "fund provider can't be empty",
			},
			{
				name:       "reject-empty-amount",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceVnd,
					domain.FundProviderTypeBank,
				),
				allocationAmount: shared.Money{},
				wantErr:          "amount can't be empty",
			},
			{
				name:       "reject-amount-does-not-match-with-wallet-currency",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceVnd,
					domain.FundProviderTypeBank,
				),
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), krw),
			},
			{
				name:       "reject-fund-provider-does-not-match-with-wallet-currenct",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceKrw,
					domain.FundProviderTypeBank,
				),
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
			},
			{
				name:       "reject-fund-provider-and-amount-does-not-match-with-wallet-currenct",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceKrw,
					domain.FundProviderTypeBank,
				),
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), krw),
			},
			{
				name:       "reject-fund-provider-already-allocated",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceVnd,
					domain.FundProviderTypeBank,
				),
				isFundProviderAllocated: true,
				allocationAmount:        assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
			},
			{
				name:       "reject-when-fund-provider-and-wallet-does-not-belong-in-the-same-office",
				officeUUID: models.OfficeUUID{UUID: common.NewUUIDv7()},
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceKrw,
					domain.FundProviderTypeBank,
				),
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
			},
			{
				name:       "reject-amount-exceed-fund-provider-available-balance",
				officeUUID: officeUUID,
				fundProvider: validFundProvider(
					t,
					officeUUID,
					initBalanceVnd,
					domain.FundProviderTypeBank,
				),
				allocationAmount: assertValidMoney(t, decimal.NewFromInt(1_100_000), vnd),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var allocationsData []domain.NewFundProviderAllocationData
				if tt.isFundProviderAllocated {
					allocationsData = append(allocationsData, domain.NewFundProviderAllocationData{
						FundProvider:     tt.fundProvider,
						AllocationAmount: shared.UnmarshalMoney(decimal.Zero, vnd),
					})
				}

				w, err := domain.NewWallet(
					"tai chinh tong", "",
					vnd,
					tt.officeUUID,
					allocationsData,
				)
				require.NoError(t, err)

				err = w.AllocateFundProvider(tt.fundProvider, tt.allocationAmount)
				require.Error(t, err)

				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Allocate success", func(t *testing.T) {
		fp := validFundProvider(
			t,
			officeUUID,
			initBalanceVnd,
			domain.FundProviderTypeBank,
		)

		w, err := domain.NewWallet(
			"tai chinh tong",
			"",
			vnd,
			officeUUID,
			nil,
		)
		require.NoError(t, err)

		allocationAmount := assertValidMoney(t, decimal.NewFromInt(200_000), vnd)

		err = w.AllocateFundProvider(fp, allocationAmount)
		require.NoError(t, err)

		assert.True(t, w.Balance().Equal(allocationAmount))
		assert.True(t, fp.AvailableBalance().Equal(assertValidMoney(t, decimal.NewFromInt(800_000), vnd))) // initBalance - allocationAmount = 800_000
	})

	t.Run("Allocate success with zero money", func(t *testing.T) {
		fp := validFundProvider(
			t,
			officeUUID,
			initBalanceVnd,
			domain.FundProviderTypeBank,
		)

		w, err := domain.NewWallet(
			"tai chinh tong",
			"",
			vnd,
			officeUUID,
			nil,
		)
		require.NoError(t, err)

		zeroAmount := assertValidMoney(t, decimal.Zero, vnd)

		err = w.AllocateFundProvider(fp, zeroAmount)
		require.NoError(t, err)

		assert.True(t, w.Balance().Equal(zeroAmount))
		assert.True(t, fp.AvailableBalance().Equal(initBalanceVnd)) // Available Money does not change
	})
}

func TestWallet_TopUp(t *testing.T) {
	krw := shared.MustNewCurrency("KRW")
	vnd := shared.MustNewCurrency("VND")

	initBalanceVnd := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
	bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name       string
			allocateFn func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money)
			amount     shared.Money
			fpUUID     domain.FundProviderUUID
			wantErr    string
		}{
			{
				name:   "reject-empty-amount",
				amount: shared.Money{},
				allocateFn: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				wantErr: "money for top up can't be empty",
			},
			{
				name:   "reject-zero-amount",
				amount: assertValidMoney(t, decimal.Zero, vnd),
				allocateFn: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				wantErr: "amount for top up must be positive",
			},
			{
				name:   "reject-negative-amount",
				amount: assertValidMoney(t, decimal.NewFromInt(-100_000), vnd),
				allocateFn: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				wantErr: "amount for top up must be positive",
			},
			{
				name: "reject-amount-currency-mismatch",
				allocateFn: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount: assertValidMoney(t, decimal.NewFromInt(100_000), krw),
			},
			{
				name:   "reject-unregistered-fund-provider",
				amount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

				fp, err := domain.NewFundProvider("Techcombank-Bach", officeUUID, domain.FundProviderTypeBank, initBalanceVnd, bankMetadata)
				require.NoError(t, err)

				w, err := domain.NewWallet("tai chinh tong", "", vnd, officeUUID, nil)
				require.NoError(t, err)

				if tt.allocateFn != nil {
					tt.allocateFn(t, w, fp, assertValidMoney(t, decimal.NewFromInt(200_000), vnd))
				}

				_, err = w.TopUp(tt.amount, fp.UUID())

				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("top-up-success", func(t *testing.T) {
		bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")
		vnd := shared.MustNewCurrency("VND")

		initBalanceVnd := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
		officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

		fp, err := domain.NewFundProvider("Techcombank-Bach", officeUUID, domain.FundProviderTypeBank, initBalanceVnd, bankMetadata)
		require.NoError(t, err)

		w, err := domain.NewWallet("Vi Tai Chinh Tong", "", vnd, officeUUID, []domain.NewFundProviderAllocationData{
			{
				FundProvider:     fp,
				AllocationAmount: assertValidMoney(t, decimal.NewFromInt(200_000), vnd),
			},
		})
		require.NoError(t, err)

		snapshot, err := w.TopUp(assertValidMoney(t, decimal.NewFromInt(200_000), vnd), fp.UUID())
		require.NoError(t, err)

		assert.Equal(t, w.UUID(), snapshot.WalletUUID)
		assert.Equal(t, fp.UUID(), snapshot.FundProviderUUID)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(400_000), vnd), w.Balance())
		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(400_000), vnd), snapshot.WalletBalance)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(1_200_000), vnd), fp.Balance())
		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(1_200_000), vnd), snapshot.FundProviderBalance)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(400_000), vnd), snapshot.AllocationMoney)
	})
}

func TestWallet_Withdraw(t *testing.T) {
	krw := shared.MustNewCurrency("KRW")
	vnd := shared.MustNewCurrency("VND")

	initBalanceVnd := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
	bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name     string
			allocate func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money)
			amount   shared.Money
			wantErr  string
		}{
			{
				name:   "reject-empty-amount",
				amount: shared.Money{},
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				wantErr: "money for top up can't be empty",
			},
			{
				name: "reject-zero-amount",
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount:  assertValidMoney(t, decimal.Zero, vnd),
				wantErr: "amount for top up must be positive",
			},
			{
				name: "reject-negative-amount",
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount:  assertValidMoney(t, decimal.NewFromInt(-100_000), vnd),
				wantErr: "amount for top up must be positive",
			},
			{
				name: "reject-amount-currency-mismatch",
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount: assertValidMoney(t, decimal.NewFromInt(100_000), krw),
			},
			{
				name:   "reject-unregistered-fund-provider",
				amount: assertValidMoney(t, decimal.NewFromInt(100_000), vnd),
			},
			{
				name: "reject-amount-greater-than-wallet-balance",
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount:  assertValidMoney(t, decimal.NewFromInt(500_000), vnd),
				wantErr: "withdraw amount must be less then wallet balance",
			},
			{
				name: "reject-amount-greater-than-fund-provider-allocation",
				allocate: func(t *testing.T, w *domain.Wallet, fp *domain.FundProvider, amount shared.Money) {
					t.Helper()

					err := w.AllocateFundProvider(fp, amount)
					require.NoError(t, err)
				},
				amount:  assertValidMoney(t, decimal.NewFromInt(300_000), vnd),
				wantErr: "allocation amount can't exceed wallet allocation",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

				fp, err := domain.NewFundProvider("Techcombank-Bach", officeUUID, domain.FundProviderTypeBank, initBalanceVnd, bankMetadata)
				require.NoError(t, err)

				fp2, err := domain.NewFundProvider("Techcombank-Bach", officeUUID, domain.FundProviderTypeBank, initBalanceVnd, bankMetadata)
				require.NoError(t, err)

				allocationMoney := assertValidMoney(t, decimal.NewFromInt(200_000), vnd)

				w, err := domain.NewWallet("tai chinh tong", "", vnd, officeUUID, []domain.NewFundProviderAllocationData{
					{
						FundProvider:     fp2,
						AllocationAmount: allocationMoney,
					},
				})
				require.NoError(t, err)

				if tt.allocate != nil {
					tt.allocate(t, w, fp, allocationMoney)
				}

				_, err = w.Withdraw(tt.amount, fp.UUID())

				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("withdraw-success", func(t *testing.T) {
		bankMetadata := validBankMetadata(t, "7777777316", "Techcombank")
		vnd := shared.MustNewCurrency("VND")

		initBalanceVnd := assertValidMoney(t, decimal.NewFromInt(1_000_000), vnd)
		officeUUID := models.OfficeUUID{UUID: common.NewUUIDv7()}

		fp, err := domain.NewFundProvider("Techcombank-Bach", officeUUID, domain.FundProviderTypeBank, initBalanceVnd, bankMetadata)
		require.NoError(t, err)

		w, err := domain.NewWallet("Vi Tai Chinh Tong", "", vnd, officeUUID, []domain.NewFundProviderAllocationData{
			{
				FundProvider:     fp,
				AllocationAmount: assertValidMoney(t, decimal.NewFromInt(200_000), vnd),
			},
		})
		require.NoError(t, err)

		snapshot, err := w.Withdraw(assertValidMoney(t, decimal.NewFromInt(100_000), vnd), fp.UUID())
		require.NoError(t, err)

		assert.Equal(t, w.UUID(), snapshot.WalletUUID)
		assert.Equal(t, fp.UUID(), snapshot.FundProviderUUID)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(100_000), vnd), w.Balance())
		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(100_000), vnd), snapshot.WalletBalance)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(900_000), vnd), fp.Balance())
		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(900_000), vnd), snapshot.FundProviderBalance)

		assert.Equal(t, assertValidMoney(t, decimal.NewFromInt(100_000), vnd), snapshot.AllocationMoney)
	})
}

func validFundProvider(
	t *testing.T,
	officeUUID models.OfficeUUID,
	initBalance shared.Money,
	fpType domain.FundProviderType,
) *domain.FundProvider {
	t.Helper()

	var metadata domain.FundProviderMetadata
	var err error

	switch fpType {
	case domain.FundProviderTypeBank:
		metadata = validBankMetadata(t, "7777777316", "Techcombank")
	case domain.FundProviderTypeCash:
		metadata, err = domain.NewCashFundProviderMetadata("Huynh Trang")
		require.NoError(t, err)
	}

	fp, err := domain.NewFundProvider(
		"Vi tai chinh tong",
		officeUUID,
		domain.FundProviderTypeBank,
		initBalance,
		metadata,
	)
	require.NoError(t, err)

	return fp
}
