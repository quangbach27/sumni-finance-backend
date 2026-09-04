package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/treasury/domain"
)

func newBankInfoData() domain.BankInfoData {
	return domain.BankInfoData{
		Name:      "Vietcombank",
		Bin:       "970436",
		ShortName: "VCB",
	}
}

func newValidBankMetadata(t *testing.T) domain.BankMetadata {
	t.Helper()

	m, err := domain.NewBankMetadata("1234567890", "John Doe", newBankInfoData())
	require.NoError(t, err)

	return m
}

func TestNewBankMetadata_Success(t *testing.T) {
	t.Parallel()

	m, err := domain.NewBankMetadata("1234567890", "John Doe", newBankInfoData())
	require.NoError(t, err)

	b := m.BankInfo()
	assert.Equal(t, "1234567890", m.AccountNumber())
	assert.Equal(t, "John Doe", m.AccountOwner())
	assert.Equal(t, "Vietcombank", b.Name())
	assert.Equal(t, "970436", b.Bin())
	assert.Equal(t, "VCB", b.ShortName())
}

func TestNewBankMetadata_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accountNumber string
		accountOwner  string
		bankInfoData  domain.BankInfoData
		wantErr       string
	}{
		{
			name:          "reject-empty-account-number",
			accountNumber: "",
			accountOwner:  "John Doe",
			bankInfoData:  newBankInfoData(),
			wantErr:       "empty-account-number",
		},
		{
			name:          "reject-empty-account-owner",
			accountNumber: "1234567890",
			accountOwner:  "",
			bankInfoData:  newBankInfoData(),
			wantErr:       "empty-account-owner",
		},
		{
			name:          "reject-empty-bank-name",
			accountNumber: "1234567890",
			accountOwner:  "John Doe",
			bankInfoData: domain.BankInfoData{
				Name:      "",
				Bin:       "970436",
				ShortName: "VCB",
			},
			wantErr: "name is required",
		},
		{
			name:          "reject-empty-bin",
			accountNumber: "1234567890",
			accountOwner:  "John Doe",
			bankInfoData: domain.BankInfoData{
				Name:      "Vietcombank",
				Bin:       "",
				ShortName: "VCB",
			},
			wantErr: "bin is required",
		},
		{
			name:          "reject-empty-short-name",
			accountNumber: "1234567890",
			accountOwner:  "John Doe",
			bankInfoData: domain.BankInfoData{
				Name:      "Vietcombank",
				Bin:       "970436",
				ShortName: "",
			},
			wantErr: "short name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewBankMetadata(tt.accountNumber, tt.accountOwner, tt.bankInfoData)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBankInfo_Value(t *testing.T) {
	t.Parallel()

	m := newValidBankMetadata(t)
	b := m.BankInfo()

	val, err := b.Value()
	require.NoError(t, err)

	data, ok := val.(string)
	require.True(t, ok)
	assert.JSONEq(t, `{
		"name": "Vietcombank",
		"bin": "970436",
		"short_name": "VCB"
	}`, data)
}

func TestBankInfo_Scan(t *testing.T) {
	t.Run("scans from string", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan(`{"name":"Vietcombank","bin":"970436","short_name":"VCB"}`)
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, "970436", b.Bin())
		assert.Equal(t, "VCB", b.ShortName())
	})

	t.Run("scans from bytes", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan([]byte(`{"name":"Vietcombank","bin":"970436","short_name":"VCB"}`))
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, "970436", b.Bin())
		assert.Equal(t, "VCB", b.ShortName())
	})

	t.Run("returns error on invalid json", func(t *testing.T) {
		t.Parallel()

		var b domain.BankInfo
		err := b.Scan("not-valid-json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}

func TestFundSourceBankMetadata_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()
		var m domain.BankMetadata
		assert.True(t, m.IsZero())
	})

	t.Run("non-zero value", func(t *testing.T) {
		t.Parallel()
		m := newValidBankMetadata(t)
		assert.False(t, m.IsZero())
	})
}

func TestFundSourceBankMetadata_MatchesType(t *testing.T) {
	t.Parallel()

	m := newValidBankMetadata(t)

	assert.True(t, m.MatchesType(domain.FundSourceTypeBank))
	assert.False(t, m.MatchesType(domain.FundSourceTypeCash))
}

func TestNewCashMetadata_Success(t *testing.T) {
	t.Parallel()

	m, err := domain.NewCashMetadata("Jane Doe")
	require.NoError(t, err)

	assert.Equal(t, "Jane Doe", m.OwnerName())
}

func TestNewCashMetadata_ValidationErrors(t *testing.T) {
	t.Parallel()

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

			_, err := domain.NewCashMetadata(tt.ownerName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestFundSourceCashMetadata_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()
		var m domain.CashMetadata
		assert.True(t, m.IsZero())
	})

	t.Run("non-zero value", func(t *testing.T) {
		t.Parallel()
		m, err := domain.NewCashMetadata("Jane Doe")
		require.NoError(t, err)
		assert.False(t, m.IsZero())
	})
}

func TestFundSourceCashMetadata_MatchesType(t *testing.T) {
	t.Parallel()

	m, err := domain.NewCashMetadata("Jane Doe")
	require.NoError(t, err)

	assert.True(t, m.MatchesType(domain.FundSourceTypeCash))
	assert.False(t, m.MatchesType(domain.FundSourceTypeBank))
}
