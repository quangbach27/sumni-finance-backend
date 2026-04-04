package fundprovider_test

import (
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFundProvider_NewFundProvider(t *testing.T) {
	testCases := []struct {
		name              string
		fpName            string
		fpType            string
		initBalanceAmount int64
		currencyCode      string
		hasErr            bool
		expectedErr       string
	}{
		{
			name:              "returns errors when fund provider name is empty",
			fpName:            "",
			fpType:            "BANK",
			initBalanceAmount: 100,
			currencyCode:      "USD",
			hasErr:            true,
		},
		{
			name:              "returns error when init balance is negative",
			fpName:            "Techcombank7316",
			fpType:            "BANK",
			initBalanceAmount: -100,
			currencyCode:      "USD",
			hasErr:            true,
		},
		{
			name:              "returns error when type is empty",
			fpName:            "Techcombank7316",
			fpType:            "",
			initBalanceAmount: 100,
			currencyCode:      "USD",
			hasErr:            true,
		},
		{
			name:              "returns error when type is invalid",
			fpName:            "Techcombank7316",
			fpType:            "INVALID",
			initBalanceAmount: 100,
			currencyCode:      "USD",
			hasErr:            true,
		},
		{
			name:              "create fund provider when init balance is zero",
			fpName:            "Techcombank7316",
			fpType:            "BANK",
			initBalanceAmount: 0,
			currencyCode:      "USD",
			hasErr:            false,
		},
		{
			name:              "create fund provider when init balance is positive",
			fpName:            "Techcombank7316",
			fpType:            "BANK",
			initBalanceAmount: 100,
			currencyCode:      "USD",
			hasErr:            false,
		},
		{
			name:              "create fund provider with CASH type",
			fpName:            "Office Cash",
			fpType:            "CASH",
			initBalanceAmount: 5000,
			currencyCode:      "USD",
			hasErr:            false,
		},
		{
			name:              "returns error when currency code is empty",
			fpName:            "Techcombank7316",
			fpType:            "BANK",
			initBalanceAmount: 0,
			currencyCode:      "",
			hasErr:            true,
		},
		{
			name:              "returns error when currency code is invalid",
			fpName:            "Techcombank7316",
			fpType:            "BANK",
			initBalanceAmount: 0,
			currencyCode:      "INVALID",
			hasErr:            true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fundProvider, err := fundprovider.NewFundProvider(tt.fpName, tt.fpType, tt.initBalanceAmount, tt.currencyCode)

			if tt.hasErr {
				require.Error(t, err)
				assert.Nil(t, fundProvider)
			} else {
				require.NoError(t, err)

				assert.Equal(t, fundProvider.Balance().Amount(), tt.initBalanceAmount)
				assert.Equal(t, fundProvider.Balance().Currency().Code(), tt.currencyCode)
			}
		})
	}
}

func TestFundProvider_UnmarshallFromDatabase(t *testing.T) {
	testCases := []struct {
		name string

		// Given
		id                       uuid.UUID
		fpName                   string
		balanceAmount            int64
		unallocatedBalanceAmount int64
		currencyCode             string
		version                  int32

		// Then
		hasErr bool
	}{
		{
			name:                     "returns error when id is empty",
			id:                       uuid.UUID{},
			fpName:                   "Techcombank7316",
			balanceAmount:            50,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "returns error when fund provider name is empty",
			id:                       uuid.New(),
			fpName:                   "",
			balanceAmount:            50,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "returns error when balance amount is negative",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            -100,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "unmarshal fund provider successfully when balance amount is zero",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            0,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   false,
		},
		{
			name:                     "unmarshal fund provider successfully when balance amount is positive",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   false,
		},
		{
			name:                     "returns error when unallocated amount is negative",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: -10,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "returns error when unallocated amount exceed balance amount",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 120,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "unmarshal fund provider successfully when unallocated amount is zero",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 0,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   false,
		},
		{
			name:                     "unmarshal fund provider successfully when unallocated amount does not excced balance",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 50,
			currencyCode:             "USD",
			version:                  0,
			hasErr:                   false,
		},
		{
			name:                     "returns error when currency code is empty",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 50,
			currencyCode:             "",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "returns error when currency code is invalid",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 50,
			currencyCode:             "INVALID",
			version:                  0,
			hasErr:                   true,
		},
		{
			name:                     "unmarshal fund provider successfully",
			id:                       uuid.New(),
			fpName:                   "Techcombank7316",
			balanceAmount:            100,
			unallocatedBalanceAmount: 50,
			currencyCode:             "USD",
			version:                  1,
			hasErr:                   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fundProvider, err := fundprovider.UnmarshalFundProviderFromDatabase(
				tt.id,
				tt.fpName,
				"BANK",
				tt.balanceAmount,
				tt.unallocatedBalanceAmount,
				tt.currencyCode,
				tt.version,
			)

			if tt.hasErr {
				require.Error(t, err)
				assert.Nil(t, fundProvider)
			} else {
				require.NoError(t, err)

				assert.Equal(t, tt.id, fundProvider.ID())

				assert.Equal(t, tt.balanceAmount, fundProvider.Balance().Amount())
				assert.Equal(t, tt.currencyCode, fundProvider.Balance().Currency().Code())

				assert.Equal(t, tt.unallocatedBalanceAmount, fundProvider.UnallocatedBalance().Amount())
				assert.Equal(t, tt.currencyCode, fundProvider.UnallocatedBalance().Currency().Code())

				assert.Equal(t, tt.version, fundProvider.Version())
			}
		})
	}
}

func TestFundProvider_Reserve(t *testing.T) {
	testCases := []struct {
		name              string
		unallocatedAmount int64
		allocatedAmount   int64
		hasErr            bool
	}{
		{
			name:              "returns error when allocated amount exceed unallocated amount",
			unallocatedAmount: 50,
			allocatedAmount:   100,
			hasErr:            true,
		},
		{
			name:              "returns error when allocated amount is negative",
			unallocatedAmount: 50,
			allocatedAmount:   -10,
			hasErr:            true,
		},
		{
			name:              "reverses unallocated amount successfully when allocated is zero",
			unallocatedAmount: 50,
			allocatedAmount:   0,
			hasErr:            false,
		},
		{
			name:              "reverses unallocated amount successfully",
			unallocatedAmount: 50,
			allocatedAmount:   10,
			hasErr:            false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given
			fundProvider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", tt.unallocatedAmount, "USD")
			require.NoError(t, err)
			require.NotNil(t, fundProvider)

			allocated, err := valueobject.NewMoney(tt.allocatedAmount, valueobject.USD)
			require.NoError(t, err)

			// When
			err = fundProvider.Reserve(allocated)

			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				expectedUnallocated := tt.unallocatedAmount - tt.allocatedAmount
				assert.Equal(t, expectedUnallocated, fundProvider.UnallocatedBalance().Amount())
			}
		})
	}
}

func TestFundProvider_TopUp(t *testing.T) {
	testCases := []struct {
		name        string
		initBalance int64
		topUpAmount int64
		hasErr      bool
		expectErr   error
	}{
		{
			name:        "returns error when amount is negative",
			initBalance: 100,
			topUpAmount: -50,
			hasErr:      true,
			expectErr:   fundprovider.ErrInsufficientAmount,
		},
		{
			name:        "returns error when amount is zero",
			initBalance: 100,
			topUpAmount: 0,
			hasErr:      true,
			expectErr:   fundprovider.ErrInsufficientAmount,
		},
		{
			name:        "top up successfully",
			initBalance: 100,
			topUpAmount: 50,
			hasErr:      false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fundProvider, err := fundprovider.NewFundProvider("Techcombank7316", "BANK", tt.initBalance, "USD")
			require.NoError(t, err)

			topUpAmount, err := valueobject.NewMoney(tt.topUpAmount, fundProvider.Currency())
			require.NoError(t, err)

			err = fundProvider.TopUp(topUpAmount)

			if tt.hasErr {
				require.Error(t, err)
				assert.ErrorIs(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)

				expectBalance := tt.initBalance + tt.topUpAmount
				assert.Equal(t, expectBalance, fundProvider.Balance().Amount())
			}
		})
	}
}

func TestFundProvider_Withdraw(t *testing.T) {
	testCases := []struct {
		name               string
		balance            int64
		unallocatedBalance int64
		withdrawAmount     int64
		hasErr             bool
		currencyTopUp      valueobject.Currency
		expectErr          error
	}{
		{
			name:               "return error when amount is negative",
			balance:            100,
			unallocatedBalance: 50,
			withdrawAmount:     -50,
			hasErr:             true,
			expectErr:          fundprovider.ErrInsufficientAmount,
		},
		{
			name:               "return error when amount is zero",
			balance:            100,
			unallocatedBalance: 50,
			withdrawAmount:     0,
			hasErr:             true,
			expectErr:          fundprovider.ErrInsufficientAmount,
		},
		{
			name:               "return error when amount excceed allocated balance",
			balance:            100,
			unallocatedBalance: 50, // allocated amount is balance - unallocatedBalance = 50
			withdrawAmount:     60,
			hasErr:             true,
		},
		{
			name:               "return error when topup currency is different from fund provider",
			balance:            100,
			unallocatedBalance: 50,
			withdrawAmount:     30,
			currencyTopUp:      valueobject.KRW,
			hasErr:             true,
		},
		{
			name:               "withdraw successfully",
			balance:            100,
			unallocatedBalance: 50,
			withdrawAmount:     30,
			hasErr:             false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fundProvider, err := fundprovider.UnmarshalFundProviderFromDatabase(
				uuid.New(),
				"Techcombank7316",
				"BANK",
				tt.balance,
				tt.unallocatedBalance,
				"USD",
				0,
			)
			require.NoError(t, err)

			var withdrawCurrency valueobject.Currency
			if tt.currencyTopUp.IsZero() {
				withdrawCurrency = fundProvider.Currency()
			} else {
				withdrawCurrency = tt.currencyTopUp
			}

			withdrawAmount, err := valueobject.NewMoney(tt.withdrawAmount, withdrawCurrency)
			require.NoError(t, err)

			err = fundProvider.Withdraw(withdrawAmount)

			if tt.hasErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.ErrorIs(t, tt.expectErr, err)
				}
			} else {
				require.NoError(t, err)

				expectBalance := tt.balance - tt.withdrawAmount
				assert.Equal(t, expectBalance, fundProvider.Balance().Amount())
			}
		})
	}
}
