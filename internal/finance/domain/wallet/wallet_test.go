package wallet_test

import (
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"sumni-finance-backend/internal/finance/domain/wallet"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWallet(t *testing.T) {
	testCases := []struct {
		name         string
		walletName   string
		currencyCode string
		hasErr       bool
	}{
		{
			name:   "returns error when name is missing",
			hasErr: true,
		},
		{
			name:         "returns error when currency code is empty",
			walletName:   "Tai chinh tong",
			currencyCode: "",
			hasErr:       true,
		},
		{
			name:         "returns error when currency code is invalid",
			walletName:   "Tai chinh tong",
			currencyCode: "INVALID",
			hasErr:       true,
		},
		{
			name:         "creates wallet successfully",
			walletName:   "Tai chinh tong",
			currencyCode: "USD",
			hasErr:       false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := wallet.NewWallet(tt.currencyCode, tt.walletName)

			if tt.hasErr {
				require.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				require.NoError(t, err)

				var expectedBalance int64 = 0
				assert.Equal(t, expectedBalance, wallet.Balance().Amount())
				assert.Equal(t, tt.currencyCode, wallet.Currency().Code())
			}
		})
	}
}
func TestUnmarshalWalletFromDatabase(t *testing.T) {
	testCases := []struct {
		name          string
		id            uuid.UUID
		walletName    string
		balanceAmount int64
		currencyCode  string
		hasErr        bool
	}{
		{
			name:   "returns error when id is empty",
			id:     uuid.UUID{},
			hasErr: true,
		},
		{
			name:       "returns error when name is empty",
			id:         uuid.New(),
			walletName: "",
			hasErr:     true,
		},
		{
			name:          "returns error when balance is negative",
			id:            uuid.New(),
			walletName:    "Tai chinh tong",
			balanceAmount: -10,
			currencyCode:  "USD",
			hasErr:        true,
		},
		{
			name:          "unmarshal wallet successfully when balance is zero",
			id:            uuid.New(),
			walletName:    "Tai chinh tong",
			balanceAmount: 0,
			currencyCode:  "USD",
			hasErr:        false,
		},
		{
			name:          "returns error when currency code is empty",
			id:            uuid.New(),
			walletName:    "Tai chinh tong",
			balanceAmount: 0,
			currencyCode:  "",
			hasErr:        true,
		},
		{
			name:          "return error when currency code is invalid",
			id:            uuid.New(),
			walletName:    "Tai chinh tong",
			balanceAmount: 0,
			currencyCode:  "INVALID",
			hasErr:        true,
		},
		{
			name:          "unmarshal wallet successfully",
			id:            uuid.New(),
			walletName:    "Tai chinh tong",
			balanceAmount: 10,
			currencyCode:  "USD",
			hasErr:        false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var version int32 = 1
			wallet, err := wallet.UnmarshalWalletFromDatabase(
				tt.id,
				tt.walletName,
				tt.balanceAmount,
				tt.currencyCode,
				version,
			)

			if tt.hasErr {
				require.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				require.NoError(t, err)

				assert.Equal(t, tt.id, wallet.ID())
				assert.Equal(t, tt.balanceAmount, wallet.Balance().Amount())
				assert.Equal(t, tt.currencyCode, wallet.Currency().Code())
				assert.Equal(t, version, wallet.Version())
			}
		})
	}
}

func TestWallet_AllocateFromFundProvider(t *testing.T) {
	t.Run("returns error when fund provider is already allocated", func(t *testing.T) {
		provider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", 100, "USD")
		require.NoError(t, err)

		allocationProvider, err := wallet.NewFpAllocation(provider, 50)
		require.NoError(t, err)

		walletDomain, err := wallet.UnmarshalWalletFromDatabase(
			uuid.New(),
			"Tai chinh tong",
			0,
			"USD",
			0,
			allocationProvider,
		)
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(provider, 50)

		require.Error(t, err)
		assert.ErrorIs(t, err, wallet.ErrFundProviderAlreadyRegistered)
	})

	t.Run("returns error when fund provider is nil", func(t *testing.T) {
		walletDomain, err := wallet.NewWallet("USD", "Tai chinh tong")
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(nil, 100)

		require.Error(t, err)
	})

	t.Run("returns error when allocated amount is negative", func(t *testing.T) {
		provider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", 100, "USD")
		require.NoError(t, err)

		walletDomain, err := wallet.NewWallet("USD", "Tai chinh tong")
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(provider, -100)

		require.Error(t, err)
		assert.ErrorIs(t, err, wallet.ErrAllocationAmountNegative)
	})

	t.Run("returns error when allocated amount exccedd unallocated amount of fund provider", func(t *testing.T) {
		provider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", 100, "USD")
		require.NoError(t, err)

		walletDomain, err := wallet.NewWallet("USD", "Tai chinh tong")
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(provider, 110)

		require.Error(t, err)
	})

	t.Run("unmarshal wallet successfully when allocatedAmount is zero", func(t *testing.T) {
		provider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", 100, "USD")
		require.NoError(t, err)

		unallocatedBalance := provider.UnallocatedBalance()

		walletDomain, err := wallet.NewWallet("USD", "Tai chinh tong")
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(provider, 0)

		require.NoError(t, err)

		actualAllocation, found := walletDomain.FundProviderManager().FindFundProviderAllocation(provider.ID())
		require.True(t, found)
		assert.Equal(t, actualAllocation.FundProvider().ID(), provider.ID())
		assert.Equal(t, unallocatedBalance, actualAllocation.FundProvider().UnallocatedBalance())
	})

	t.Run("unmarshal wallet successfully", func(t *testing.T) {
		provider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", 100, "USD")
		require.NoError(t, err)

		unallocatedBalance := provider.UnallocatedBalance()

		walletDomain, err := wallet.NewWallet("USD", "Tai chinh tong")
		require.NoError(t, err)

		err = walletDomain.AllocateFundProvider(provider, 50)

		require.NoError(t, err)

		actualAllocation, found := walletDomain.FundProviderManager().FindFundProviderAllocation(provider.ID())
		require.True(t, found)
		assert.Equal(t, actualAllocation.FundProvider().ID(), provider.ID())
		assert.Equal(t, unallocatedBalance.Amount()-50, actualAllocation.FundProvider().UnallocatedBalance().Amount())
	})
}
