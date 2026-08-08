package domain_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
)

var walletTestVND = shared.MustNewCurrency("VND")

func TestNewWallet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wallet   string
		currency shared.Currency
		hasErr   bool
		errHas   string
		wantName string
		wantCur  shared.Currency
		wantZero string
	}{
		{
			name:     "creates wallet successfully",
			wallet:   "Cash Wallet",
			currency: shared.MustNewCurrency("VND"),
			wantName: "Cash Wallet",
			wantCur:  shared.MustNewCurrency("VND"),
			wantZero: "0",
		},
		{
			name:     "returns error when wallet name is empty",
			wallet:   "",
			currency: shared.MustNewCurrency("VND"),
			hasErr:   true,
			errHas:   "name-empty",
		},
		{
			name:   "returns error when currency is empty",
			wallet: "Cash Wallet",
			hasErr: true,
			errHas: "currency-empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wallet, err := domain.NewWallet(tt.wallet, tt.currency)

			if tt.hasErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errHas)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, wallet)

			assert.False(t, wallet.UUID().IsZero())
			assert.Equal(t, tt.wantName, wallet.Name())
			assert.Equal(t, tt.wantZero, wallet.Balance().Amount().String())
			assert.True(t, wallet.Balance().Currency().Equal(tt.wantCur))
			assert.NotNil(t, wallet.Allocations())
			assert.Len(t, wallet.Allocations(), 0)
		})
	}
}

func TestWalet_Allocate(t *testing.T) {
	t.Parallel()

	t.Run("allocates successfully when fund source has enough available balance", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))

		allocatedAmount := newWalletTestMoney(t, 100_000, walletTestVND)
		err := wallet.Allocate(fs, allocatedAmount)
		require.NoError(t, err)

		assert.Equal(t, fs.Balance(), newWalletTestMoney(t, 1_000_000, walletTestVND))
		assert.Equal(t, fs.AvailableBalance(), newWalletTestMoney(t, 900_000, walletTestVND))
		assert.Equal(t, wallet.Allocations()[fs.UUID()].Balance(), allocatedAmount)
		assert.Equal(t, wallet.Balance(), allocatedAmount)
	})

	t.Run("returns error when allocating the same fund source twice", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))

		allocatedAmount := newWalletTestMoney(t, 100_000, walletTestVND)
		err := wallet.Allocate(fs, allocatedAmount)
		require.NoError(t, err)

		err = wallet.Allocate(fs, allocatedAmount)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already allocated")
	})

	t.Run("returns error when allocated amount exceeds available balance", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))

		err := wallet.Allocate(fs, newWalletTestMoney(t, 2_000_000, walletTestVND))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to allocated")
		assert.Contains(t, err.Error(), "insufficient available balance")

		assert.Equal(t, "0", wallet.Balance().Amount().String())
		assert.True(t, wallet.Balance().Currency().Equal(walletTestVND))
		assert.Len(t, wallet.Allocations(), 0)
	})
}

func TestWallet_TopUp(t *testing.T) {
	t.Parallel()

	t.Run("tops up wallet and fund source successfully", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newValidFundSource(t, newValidMoney(t, 1_000_000, vnd))
		allocatedAmount := newValidMoney(t, 100_000, vnd)

		err := wallet.Allocate(fs, allocatedAmount)
		require.NoError(t, err)

		err = wallet.TopUp(fs.UUID(), newValidMoney(t, 500_000, vnd))
		require.NoError(t, err)

		assert.Equal(t, wallet.Balance(), newValidMoney(t, 600_000, vnd))
		assert.Equal(t, wallet.Allocations()[fs.UUID()].Balance(), newValidMoney(t, 600_000, vnd))
		assert.Equal(t, fs.Balance(), newValidMoney(t, 1_500_000, vnd))
	})

	t.Run("returns error when amount is not positive", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newValidFundSource(t, newValidMoney(t, 1_000_000, vnd))
		err := wallet.Allocate(fs, newValidMoney(t, 100_000, vnd))
		require.NoError(t, err)

		err = wallet.TopUp(fs.UUID(), newValidMoney(t, 0, vnd))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "top up amount must be positive")
	})

	t.Run("returns error when fund source is not allocated", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newValidFundSource(t, newValidMoney(t, 1_000_000, vnd))

		err := wallet.TopUp(fs.UUID(), newValidMoney(t, 100_000, vnd))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not in wallet allocation")
	})

	t.Run("returns error when currency mismatches fund source", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newValidFundSource(t, newValidMoney(t, 1_000_000, vnd))
		err := wallet.Allocate(fs, newValidMoney(t, 100_000, vnd))
		require.NoError(t, err)

		err = wallet.TopUp(fs.UUID(), newValidMoney(t, 100_000, krw))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to top up fund source")
		assert.Contains(t, err.Error(), "currency mismatch")
	})
}

func TestWallet_Withdraw(t *testing.T) {
	t.Parallel()

	t.Run("withdraws successfully from allocated fund source", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))

		allocatedAmount := newWalletTestMoney(t, 200_000, walletTestVND)
		err := wallet.Allocate(fs, allocatedAmount)
		require.NoError(t, err)

		err = wallet.Withdraw(fs.UUID(), newWalletTestMoney(t, 50_000, walletTestVND))
		require.NoError(t, err)

		assert.Equal(t, "150000", wallet.Balance().Amount().String())
		assert.Equal(t, "150000", wallet.Allocations()[fs.UUID()].Balance().Amount().String())
		assert.Equal(t, "950000", fs.Balance().Amount().String())
	})

	t.Run("returns error when amount is not positive", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))
		err := wallet.Allocate(fs, newWalletTestMoney(t, 100_000, walletTestVND))
		require.NoError(t, err)

		err = wallet.Withdraw(fs.UUID(), newWalletTestMoney(t, 0, walletTestVND))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "withdraw amount must be positive")
	})

	t.Run("returns error when fund source is not allocated", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))

		err := wallet.Withdraw(fs.UUID(), newWalletTestMoney(t, 50_000, walletTestVND))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not in wallet allocation")
	})

	t.Run("returns error when allocation balance is insufficient", func(t *testing.T) {
		t.Parallel()

		wallet := newValidWallet(t)
		fs := newWalletTestFundSource(t, newWalletTestMoney(t, 1_000_000, walletTestVND))
		err := wallet.Allocate(fs, newWalletTestMoney(t, 100_000, walletTestVND))
		require.NoError(t, err)

		err = wallet.Withdraw(fs.UUID(), newWalletTestMoney(t, 200_000, walletTestVND))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wallet doesn't have enough money")
	})
}

func newValidWallet(t *testing.T) *domain.Wallet {
	t.Helper()

	wallet, err := domain.NewWallet(gofakeit.Company(), walletTestVND)
	require.NoError(t, err)

	return wallet
}

func newWalletTestMoney(t *testing.T, amount int64, currency shared.Currency) shared.Money {
	t.Helper()

	m, err := shared.NewMoney(decimal.NewFromInt(amount), currency)
	require.NoError(t, err)

	return m
}

func newWalletTestFundSource(t *testing.T, balance shared.Money) *domain.FundSource {
	t.Helper()

	metadata, err := domain.NewCashMetadata("John Doe")
	require.NoError(t, err)

	fs, err := domain.NewFundSource("Wallet", domain.FundSourceTypeCash, balance, walletTestVND, metadata)
	require.NoError(t, err)

	return fs
}
