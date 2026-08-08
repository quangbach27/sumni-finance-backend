package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
)

var (
	vnd = shared.MustNewCurrency("VND")
	krw = shared.MustNewCurrency("KRW")
)

// Reserve

func TestFundSource_Reserve_Success(t *testing.T) {
	t.Parallel()

	fs := newValidFundSource(t, newValidMoney(t, 500_000, vnd))

	require.NoError(t, fs.Reserve(newValidMoney(t, 200_000, vnd)))

	assert.Equal(t, "500000", fs.Balance().Amount().String())
	assert.Equal(t, "300000", fs.AvailableBalance().Amount().String())
}

func TestFundSource_Reserve_ZeroAmount(t *testing.T) {
	t.Parallel()

	fs := newValidFundSource(t, newValidMoney(t, 500_000, vnd))

	require.NoError(t, fs.Reserve(newValidMoney(t, 0, vnd)))

	assert.Equal(t, "500000", fs.AvailableBalance().Amount().String())
}

func TestFundSource_Reserve_ValidationErrors(t *testing.T) {
	t.Parallel()

	fs := newValidFundSource(t, newValidMoney(t, 500_000, vnd))

	tests := []struct {
		name    string
		amount  shared.Money
		wantErr string
	}{
		{
			name:    "reject-currency-mismatch",
			amount:  newValidMoney(t, 100, krw),
			wantErr: "does not match fund source currency",
		},
		{
			name:    "reject-negative-amount",
			amount:  newValidMoney(t, -1, vnd),
			wantErr: "reservation amount must be greater or equal than zero",
		},
		{
			name:    "reject-exceeds-available-balance",
			amount:  newValidMoney(t, 600_000, vnd),
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

type newValidFundSourceOptions struct {
	fsType domain.FundSourceType
}

func newValidFundSource(t *testing.T, balance shared.Money, opts ...newValidFundSourceOptions) *domain.FundSource {
	t.Helper()

	fsType := domain.FundSourceTypeCash
	if len(opts) > 0 && !opts[0].fsType.IsZero() {
		fsType = opts[0].fsType
	}

	var (
		metadata domain.FundSourceMetadata
		err      error
	)

	switch fsType {
	case domain.FundSourceTypeBank:
		metadata, err = domain.NewBankMetadata("1234567890", "John Doe", domain.BankInfoData{
			Name:      "Vietcombank",
			Bin:       "970436",
			ShortName: "VCB",
		})
	default:
		metadata, err = domain.NewCashMetadata("John Doe")
	}
	require.NoError(t, err)

	fs, err := domain.NewFundSource("Wallet", fsType, balance, vnd, metadata)
	require.NoError(t, err)

	return fs
}

func newValidMoney(t *testing.T, amount int64, currency shared.Currency) shared.Money {
	t.Helper()
	m, err := shared.NewMoney(decimal.NewFromInt(amount), currency)
	require.NoError(t, err)
	return m
}
