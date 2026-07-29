package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/treasury/adapters/bankprovider"
	"sumni-finance-backend/internal/treasury/domain"
)

func newTestFactory(t *testing.T, withErr bool) *domain.FundSourceFactory {
	t.Helper()
	var bankProvider domain.BankProvider
	if withErr {
		bankProvider = bankprovider.NewStubClientWithError(errors.New("failed to retrive bank info"))
	} else {
		bankProvider = bankprovider.NewStubClient()
	}
	return domain.NewFundSourceFactory(bankProvider)
}

func TestNewBankInfo_Success(t *testing.T) {
	t.Parallel()

	b, err := domain.NewBankInfo("Vietcombank", "970436", "VCB", "VCB", "https://logo.url")
	require.NoError(t, err)

	assert.Equal(t, "Vietcombank", b.Name())
	assert.Equal(t, "970436", b.Bin())
	assert.Equal(t, "VCB", b.BankCode())
	assert.Equal(t, "VCB", b.ShortName())
	assert.Equal(t, "https://logo.url", b.LogoUrl())
}

func TestNewBankInfo_ValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		bankName        string
		bin             string
		bankCode        string
		shortName       string
		logoUrl         string
		iconUrl         string
		lookupSupport   int
		transferSupport int
		wantErr         string
	}{
		{
			name:      "missing-name",
			bankName:  "",
			bin:       "970436",
			bankCode:  "VCB",
			shortName: "VCB",
			logoUrl:   "https://logo.url",
			wantErr:   "name is required",
		},
		{
			name:      "missing-bin",
			bankName:  "Vietcombank",
			bin:       "",
			bankCode:  "VCB",
			shortName: "VCB",
			logoUrl:   "https://logo.url",
			wantErr:   "bin is required",
		},
		{
			name:      "missing-bank-code",
			bankName:  "Vietcombank",
			bin:       "970436",
			bankCode:  "",
			shortName: "VCB",
			logoUrl:   "https://logo.url",
			wantErr:   "bank code is required",
		},
		{
			name:      "missing-short-name",
			bankName:  "Vietcombank",
			bin:       "970436",
			bankCode:  "VCB",
			shortName: "",
			logoUrl:   "https://logo.url",
			wantErr:   "short name is required",
		},
		{
			name:      "missing-logo-url",
			bankName:  "Vietcombank",
			bin:       "970436",
			bankCode:  "VCB",
			shortName: "VCB",
			logoUrl:   "",
			wantErr:   "logo url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewBankInfo(tt.bankName, tt.bin, tt.bankCode, tt.shortName, tt.logoUrl)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBankInfo_Value(t *testing.T) {
	t.Parallel()

	b, err := domain.NewBankInfo("Vietcombank", "970436", "VCB", "VCB", "https://logo.url")
	require.NoError(t, err)

	val, err := b.Value()
	require.NoError(t, err)

	data, ok := val.(string)
	require.True(t, ok)
	assert.JSONEq(t, `{
		"name": "Vietcombank",
		"bin": "970436",
		"bankCode": "VCB",
		"shortName": "VCB",
		"logoUrl": "https://logo.url"
	}`, data)
}

func TestBankInfo_Scan(t *testing.T) {
	t.Run("scans from string", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan(`{"name":"Vietcombank","bin":"970436","bankCode":"VCB","shortName":"VCB","logoUrl":"https://logo.url"}`)
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, "970436", b.Bin())
		assert.Equal(t, "VCB", b.BankCode())
		assert.Equal(t, "VCB", b.ShortName())
		assert.Equal(t, "https://logo.url", b.LogoUrl())
	})

	t.Run("scans from bytes", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan([]byte(`{"name":"Vietcombank","bin":"970436","bankCode":"VCB","shortName":"VCB","logoUrl":"https://logo.url"}`))
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, "970436", b.Bin())
	})

	t.Run("returns error on invalid json", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan("not-valid-json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}

func TestNewFundSourceBankMetadata_Success(t *testing.T) {
	t.Parallel()

	factory := newTestFactory(t, false)
	m, err := factory.NewBankMetadata(context.Background(), "VCB", "1234567890", "John Doe")
	require.NoError(t, err)

	assert.Equal(t, "1234567890", m.AccountNumber())
	assert.Equal(t, "John Doe", m.AccountOwner())
	assert.False(t, m.BankInfo().IsZero())
}

func TestNewFundSourceBankMetadata_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		bankProviderErr bool
		bankCode        string
		accountNumber   string
		accountOwner    string
		wantErr         string
	}{
		{
			name:          "reject-empty-account-number",
			bankCode:      "VCB",
			accountNumber: "",
			accountOwner:  "John Doe",
			wantErr:       "empty-account-number",
		},
		{
			name:          "reject-empty-account-owner",
			bankCode:      "VCB",
			accountNumber: "1234567890",
			accountOwner:  "",
			wantErr:       "empty-account-owner",
		},
		{
			name:            "reject-bank-not-found",
			bankProviderErr: true,
			bankCode:        "UNKNOWN",
			accountNumber:   "1234567890",
			accountOwner:    "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := newTestFactory(t, tt.bankProviderErr)
			_, err := factory.NewBankMetadata(context.Background(), tt.bankCode, tt.accountNumber, tt.accountOwner)
			require.Error(t, err)
			if len(tt.wantErr) > 0 {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestFundSourceBankMetadata_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()
		var m domain.FundSourceBankMetadata
		assert.True(t, m.IsZero())
	})

	t.Run("non-zero value", func(t *testing.T) {
		t.Parallel()
		factory := newTestFactory(t, false)
		m, err := factory.NewBankMetadata(context.Background(), "VCB", "1234567890", "John Doe")
		require.NoError(t, err)
		assert.False(t, m.IsZero())
	})
}

func TestFundSourceBankMetadata_MatchesType(t *testing.T) {
	t.Parallel()

	factory := newTestFactory(t, false)
	m, err := factory.NewBankMetadata(context.Background(), "VCB", "1234567890", "John Doe")
	require.NoError(t, err)

	assert.True(t, m.MatchesType(domain.FundSourceTypeBank))
	assert.False(t, m.MatchesType(domain.FundSourceTypeCash))
}

func TestNewFundSourceCashMetadata_Success(t *testing.T) {
	t.Parallel()

	factory := newTestFactory(t, false)
	m, err := factory.NewCashMetadata("Jane Doe")
	require.NoError(t, err)

	assert.Equal(t, "Jane Doe", m.OwnerName())
}

func TestNewFundSourceCashMetadata_ValidationErrors(t *testing.T) {
	t.Parallel()

	factory := newTestFactory(t, false)

	tests := []struct {
		name      string
		ownerName string
		wantErr   string
	}{
		{
			name:      "reject-empty-owner-name",
			ownerName: "",
			wantErr:   "empty-owner-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := factory.NewCashMetadata(tt.ownerName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestFundSourceCashMetadata_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()
		var m domain.FundSourceCashMetadata
		assert.True(t, m.IsZero())
	})

	t.Run("non-zero value", func(t *testing.T) {
		t.Parallel()
		factory := newTestFactory(t, false)
		m, err := factory.NewCashMetadata("Jane Doe")
		require.NoError(t, err)
		assert.False(t, m.IsZero())
	})
}

func TestFundSourceCashMetadata_MatchesType(t *testing.T) {
	t.Parallel()

	factory := newTestFactory(t, false)
	m, err := factory.NewCashMetadata("Jane Doe")
	require.NoError(t, err)

	assert.True(t, m.MatchesType(domain.FundSourceTypeCash))
	assert.False(t, m.MatchesType(domain.FundSourceTypeBank))
}
