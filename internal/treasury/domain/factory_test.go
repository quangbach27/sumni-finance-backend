package domain_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/adapters/bank/lookup"
	"sumni-finance-backend/internal/treasury/domain"
)

var vnd = shared.MustNewCurrency("VND")

func TestNewFundSource_BankSuccess(t *testing.T) {
	t.Parallel()

	factory := newFactory(t)
	balance := newZeroBalance(t)
	metadata := newValidBankFundMetadata(t, factory)

	fs, err := factory.NewFundSource("My Bank Account", domain.FundSourceTypeBank, balance, vnd, metadata, "user-1")
	require.NoError(t, err)

	assert.Equal(t, "My Bank Account", fs.Name())
	assert.Equal(t, domain.FundSourceTypeBank, fs.SourceType())
	assert.True(t, balance.Equal(fs.Balance()))
	assert.True(t, balance.Equal(fs.AvailableBalance()))
	assert.True(t, vnd.Equal(fs.Currency()))
	assert.False(t, fs.UUID().IsZero())
}

func TestNewFundSource_CashSuccess(t *testing.T) {
	t.Parallel()

	factory := newFactory(t)
	balance := newZeroBalance(t)
	metadata := newValidCashFundMetadata(t, factory)

	fs, err := factory.NewFundSource("My Cash Wallet", domain.FundSourceTypeCash, balance, vnd, metadata, "user-1")
	require.NoError(t, err)

	assert.Equal(t, "My Cash Wallet", fs.Name())
	assert.Equal(t, domain.FundSourceTypeCash, fs.SourceType())
}

func TestNewFundSource_ValidationErrors(t *testing.T) {
	t.Parallel()

	factory := newFactory(t)
	balance := newZeroBalance(t)
	bankMetadata := newValidBankFundMetadata(t, factory)

	negativeBalance, err := shared.NewMoney(decimal.NewFromInt(-100), vnd)
	require.NoError(t, err)

	tests := []struct {
		name        string
		fsName      string
		sourceType  domain.FundSourceType
		balance     shared.Money
		currency    shared.Currency
		metadata    domain.FundSourceMetadata
		wantErrSlug string
	}{
		{
			name:        "reject-empty-name",
			fsName:      "",
			sourceType:  domain.FundSourceTypeBank,
			balance:     balance,
			currency:    vnd,
			metadata:    bankMetadata,
			wantErrSlug: "empty-name",
		},
		{
			name:        "reject-empty-source-type",
			fsName:      "My Account",
			sourceType:  domain.FundSourceType{},
			balance:     balance,
			currency:    vnd,
			metadata:    bankMetadata,
			wantErrSlug: "empty-source-type",
		},
		{
			name:        "reject-empty-currency",
			fsName:      "My Account",
			sourceType:  domain.FundSourceTypeBank,
			balance:     balance,
			currency:    shared.Currency{},
			metadata:    bankMetadata,
			wantErrSlug: "empty-currency",
		},
		{
			name:        "reject-negative-balance",
			fsName:      "My Account",
			sourceType:  domain.FundSourceTypeBank,
			balance:     negativeBalance,
			currency:    vnd,
			metadata:    bankMetadata,
			wantErrSlug: "negative-balance",
		},
		{
			name:        "reject-nil-metadata",
			fsName:      "My Account",
			sourceType:  domain.FundSourceTypeBank,
			balance:     balance,
			currency:    vnd,
			metadata:    nil,
			wantErrSlug: "empty-metadata",
		},
		{
			name:        "reject-metadata-type-mismatch",
			fsName:      "My Account",
			sourceType:  domain.FundSourceTypeCash,
			balance:     balance,
			currency:    vnd,
			metadata:    bankMetadata,
			wantErrSlug: "metadata-type-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := factory.NewFundSource(tt.fsName, tt.sourceType, tt.balance, tt.currency, tt.metadata, "user-1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrSlug)
		})
	}
}

func newFactory(t *testing.T) *domain.FundSourceFactory {
	t.Helper()
	return domain.NewFundSourceFactory(lookup.NewStubClient())
}

func newValidBankFundMetadata(t *testing.T, factory *domain.FundSourceFactory) domain.FundSourceBankMetadata {
	t.Helper()
	m, err := factory.NewBankMetadata(context.Background(), "VCB", "1234567890", "John Doe")
	require.NoError(t, err)
	return m
}

func newValidCashFundMetadata(t *testing.T, factory *domain.FundSourceFactory) domain.FundSourceCashMetadata {
	t.Helper()
	m, err := factory.NewCashMetadata("John Doe")
	require.NoError(t, err)
	return m
}

func newZeroBalance(t *testing.T) shared.Money {
	t.Helper()
	m, err := shared.NewMoney(decimal.Zero, vnd)
	require.NoError(t, err)
	return m
}
