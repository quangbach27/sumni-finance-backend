package shared_test

import (
	"testing"

	"sumni-finance-backend/internal/common/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidBankInfo(t *testing.T) shared.BankInfo {
	t.Helper()
	b, err := shared.NewBankInfo("Vietcombank", 970436, "VCB", "VCB", "https://logo.url", "https://icon.url", 1, 1)
	require.NoError(t, err)
	return b
}

func TestNewBankInfo_Success(t *testing.T) {
	t.Parallel()

	b, err := shared.NewBankInfo("Vietcombank", 970436, "VCB", "VCB", "https://logo.url", "https://icon.url", 1, 1)
	require.NoError(t, err)

	assert.Equal(t, "Vietcombank", b.Name())
	assert.Equal(t, 970436, b.Bin())
	assert.Equal(t, "VCB", b.BankCode())
	assert.Equal(t, "VCB", b.ShortName())
	assert.Equal(t, "https://logo.url", b.LogoUrl())
	assert.Equal(t, "https://icon.url", b.IconUrl())
	assert.Equal(t, 1, b.LookupSupport())
	assert.Equal(t, 1, b.TransferSupport())
}

func TestNewBankInfo_ValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		bankName        string
		bin             int
		bankCode        string
		shortName       string
		logoUrl         string
		iconUrl         string
		lookupSupport   int
		transferSupport int
		wantErr         string
	}{
		{
			name:            "missing-name",
			bankName:        "",
			bin:             970436,
			bankCode:        "VCB",
			shortName:       "VCB",
			logoUrl:         "https://logo.url",
			iconUrl:         "https://icon.url",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "name is required",
		},
		{
			name:            "missing-bin",
			bankName:        "Vietcombank",
			bin:             0,
			bankCode:        "VCB",
			shortName:       "VCB",
			logoUrl:         "https://logo.url",
			iconUrl:         "https://icon.url",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "bin is required",
		},
		{
			name:            "missing-bank-code",
			bankName:        "Vietcombank",
			bin:             970436,
			bankCode:        "",
			shortName:       "VCB",
			logoUrl:         "https://logo.url",
			iconUrl:         "https://icon.url",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "bank code is required",
		},
		{
			name:            "missing-short-name",
			bankName:        "Vietcombank",
			bin:             970436,
			bankCode:        "VCB",
			shortName:       "",
			logoUrl:         "https://logo.url",
			iconUrl:         "https://icon.url",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "short name is required",
		},
		{
			name:            "missing-logo-url",
			bankName:        "Vietcombank",
			bin:             970436,
			bankCode:        "VCB",
			shortName:       "VCB",
			logoUrl:         "",
			iconUrl:         "https://icon.url",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "logo url is required",
		},
		{
			name:            "missing-icon-url",
			bankName:        "Vietcombank",
			bin:             970436,
			bankCode:        "VCB",
			shortName:       "VCB",
			logoUrl:         "https://logo.url",
			iconUrl:         "",
			lookupSupport:   1,
			transferSupport: 1,
			wantErr:         "icon url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := shared.NewBankInfo(tt.bankName, tt.bin, tt.bankCode, tt.shortName, tt.logoUrl, tt.iconUrl, tt.lookupSupport, tt.transferSupport)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestBankInfo_Value(t *testing.T) {
	t.Parallel()

	b := newValidBankInfo(t)

	val, err := b.Value()
	require.NoError(t, err)

	data, ok := val.(string)
	require.True(t, ok)
	assert.JSONEq(t, `{
		"name": "Vietcombank",
		"bin": 970436,
		"bankCode": "VCB",
		"shortName": "VCB",
		"logoUrl": "https://logo.url",
		"iconUrl": "https://icon.url",
		"lookupSupport": 1,
		"transferSupport": 1
	}`, data)
}

func TestBankInfo_Scan(t *testing.T) {
	t.Run("scans from string", func(t *testing.T) {
		t.Parallel()

		var b shared.BankInfo
		err := b.Scan(`{"name":"Vietcombank","bin":970436,"bankCode":"VCB","shortName":"VCB","logoUrl":"https://logo.url","iconUrl":"https://icon.url","lookupSupport":1,"transferSupport":1}`)
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, 970436, b.Bin())
		assert.Equal(t, "VCB", b.BankCode())
		assert.Equal(t, "VCB", b.ShortName())
		assert.Equal(t, "https://logo.url", b.LogoUrl())
		assert.Equal(t, "https://icon.url", b.IconUrl())
		assert.Equal(t, 1, b.LookupSupport())
		assert.Equal(t, 1, b.TransferSupport())
	})

	t.Run("scans from bytes", func(t *testing.T) {
		t.Parallel()

		var b shared.BankInfo
		err := b.Scan([]byte(`{"name":"Vietcombank","bin":970436,"bankCode":"VCB","shortName":"VCB","logoUrl":"https://logo.url","iconUrl":"https://icon.url","lookupSupport":1,"transferSupport":1}`))
		require.NoError(t, err)

		assert.Equal(t, "Vietcombank", b.Name())
		assert.Equal(t, 970436, b.Bin())
	})

	t.Run("returns error on unsupported type", func(t *testing.T) {
		t.Parallel()

		var b shared.BankInfo
		err := b.Scan(12345)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("returns error on invalid json", func(t *testing.T) {
		t.Parallel()

		var b shared.BankInfo
		err := b.Scan("not-valid-json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}

func TestBankInfo_ValueScanRoundtrip(t *testing.T) {
	t.Parallel()

	original := newValidBankInfo(t)

	val, err := original.Value()
	require.NoError(t, err)

	var restored shared.BankInfo
	err = restored.Scan(val)
	require.NoError(t, err)

	assert.Equal(t, original.Name(), restored.Name())
	assert.Equal(t, original.Bin(), restored.Bin())
	assert.Equal(t, original.BankCode(), restored.BankCode())
	assert.Equal(t, original.ShortName(), restored.ShortName())
	assert.Equal(t, original.LogoUrl(), restored.LogoUrl())
	assert.Equal(t, original.IconUrl(), restored.IconUrl())
	assert.Equal(t, original.LookupSupport(), restored.LookupSupport())
	assert.Equal(t, original.TransferSupport(), restored.TransferSupport())
}
