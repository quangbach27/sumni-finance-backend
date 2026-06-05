package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
)

var krw = shared.MustNewCurrency("KRW")

func newFundSource(t *testing.T, balance shared.Money) *domain.FundSource {
	t.Helper()
	factory := newFactory(t)
	metadata := newValidCashFundMetadata(t, factory)
	fs, err := factory.NewFundSource("Wallet", domain.FundSourceTypeCash, balance, vnd, metadata)
	require.NoError(t, err)
	return fs
}

func money(t *testing.T, amount int64, currency shared.Currency) shared.Money {
	t.Helper()
	m, err := shared.NewMoney(decimal.NewFromInt(amount), currency)
	require.NoError(t, err)
	return m
}

// TopUp

func TestFundSource_TopUp_Success(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, newZeroBalance(t))

	require.NoError(t, fs.TopUp(money(t, 100_000, vnd)))

	assert.Equal(t, "100000", fs.Balance().Amount().String())
}

func TestFundSource_TopUp_AccumulatesBalance(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, newZeroBalance(t))

	require.NoError(t, fs.TopUp(money(t, 100_000, vnd)))
	require.NoError(t, fs.TopUp(money(t, 50_000, vnd)))

	assert.Equal(t, "150000", fs.Balance().Amount().String())
}

func TestFundSource_TopUp_ValidationErrors(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, newZeroBalance(t))

	tests := []struct {
		name    string
		amount  shared.Money
		wantErr string
	}{
		{
			name:    "reject-zero-amount",
			amount:  money(t, 0, vnd),
			wantErr: "top up amount must be positive",
		},
		{
			name:    "reject-negative-amount",
			amount:  money(t, -1, vnd),
			wantErr: "top up amount must be positive",
		},
		{
			name:    "reject-currency-mismatch",
			amount:  money(t, 100, krw),
			wantErr: "currency mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := fs.TopUp(tt.amount)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// Withdraw

func TestFundSource_Withdraw_Success(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, money(t, 500_000, vnd))

	require.NoError(t, fs.Withdraw(money(t, 200_000, vnd)))

	assert.Equal(t, "300000", fs.Balance().Amount().String())
}

func TestFundSource_Withdraw_ValidationErrors(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, money(t, 500_000, vnd))

	tests := []struct {
		name    string
		amount  shared.Money
		wantErr string
	}{
		{
			name:    "reject-zero-amount",
			amount:  money(t, 0, vnd),
			wantErr: "withdraw amount must be positive",
		},
		{
			name:    "reject-negative-amount",
			amount:  money(t, -1, vnd),
			wantErr: "withdraw amount must be positive",
		},
		{
			name:    "reject-exceeds-balance",
			amount:  money(t, 600_000, vnd),
			wantErr: "insufficient balance to complete withdrawal",
		},
		{
			name:    "reject-currency-mismatch",
			amount:  money(t, 100, krw),
			wantErr: "currency mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := fs.Withdraw(tt.amount)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// Reserve

func TestFundSource_Reserve_Success(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, money(t, 500_000, vnd))

	require.NoError(t, fs.Reserve(money(t, 200_000, vnd)))

	assert.Equal(t, "500000", fs.Balance().Amount().String())
	assert.Equal(t, "300000", fs.AvailableBalance().Amount().String())
}

func TestFundSource_Reserve_ZeroAmount(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, money(t, 500_000, vnd))

	require.NoError(t, fs.Reserve(money(t, 0, vnd)))

	assert.Equal(t, "500000", fs.AvailableBalance().Amount().String())
}

func TestFundSource_Reserve_ValidationErrors(t *testing.T) {
	t.Parallel()

	fs := newFundSource(t, money(t, 500_000, vnd))

	tests := []struct {
		name    string
		amount  shared.Money
		wantErr string
	}{
		{
			name:    "reject-currency-mismatch",
			amount:  money(t, 100, krw),
			wantErr: "does not match fund source currency",
		},
		{
			name:    "reject-negative-amount",
			amount:  money(t, -1, vnd),
			wantErr: "reservation amount must be greater or equal than zero",
		},
		{
			name:    "reject-exceeds-available-balance",
			amount:  money(t, 600_000, vnd),
			wantErr: "insufficient available balance to complete reservation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := fs.Reserve(tt.amount)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
