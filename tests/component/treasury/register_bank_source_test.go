package treasury_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	treasuryClient "sumni-finance-backend/internal/treasury/api/http/client"
	"sumni-finance-backend/internal/treasury/domain"
)

func TestRegisterBankFundSource_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)

	fsUUID := registerBankSource(ctx, t, client, decimal.Zero)

	assert.False(t, fsUUID.IsZero())
}

func TestRegisterBankFundSource_DuplicateAccountReturns409(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)

	accountNumber := gofakeit.Numerify("##########")
	accountOwner := gofakeit.Name()
	vnd := shared.MustNewCurrency("VND")

	body := treasuryClient.RegisterFundSourceJSONRequestBody{
		Name:        gofakeit.Company(),
		SourceType:  domain.FundSourceTypeBank,
		InitBalance: decimal.NewFromInt(0),
		Currency:    vnd,
		Metadata: treasuryClient.FundSourceMetadata{
			BankMetadata: &treasuryClient.FundSourceBankMetadata{
				BankCode:      "VCB",
				AccountNumber: accountNumber,
				AccountOwner:  accountOwner,
			},
		},
	}

	first, err := client.RegisterFundSourceWithResponse(ctx, body)
	require.NoError(t, err)
	assertCorrelationID(t, first.HTTPResponse)
	require.NotNil(t, first.JSON201)

	second, err := client.RegisterFundSourceWithResponse(ctx, body)
	require.NoError(t, err)
	assertCorrelationID(t, second.HTTPResponse)
	assert.Equal(t, 409, second.StatusCode())
	assert.Nil(t, second.JSON201)
}

func TestRegisterBankFundSource_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	vnd := shared.MustNewCurrency("VND")

	validBankMetadata := &treasuryClient.FundSourceBankMetadata{
		BankCode:      "VCB",
		AccountNumber: gofakeit.Numerify("##########"),
		AccountOwner:  gofakeit.Name(),
	}

	validCashMetadata := &treasuryClient.FundSourceCashMetadata{
		OwnerName: gofakeit.Name(),
	}

	tests := []struct {
		name string
		body treasuryClient.RegisterFundSourceJSONRequestBody
	}{
		{
			name: "reject-empty-name",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        "",
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.Zero,
				Currency:    vnd,
				Metadata:    treasuryClient.FundSourceMetadata{BankMetadata: validBankMetadata},
			},
		},
		{
			name: "reject-negative-balance",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        gofakeit.Company(),
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.NewFromInt(-1),
				Currency:    vnd,
				Metadata:    treasuryClient.FundSourceMetadata{BankMetadata: validBankMetadata},
			},
		},
		{
			name: "reject-empty-source-type",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        gofakeit.Company(),
				SourceType:  domain.FundSourceType{},
				InitBalance: decimal.Zero,
				Currency:    vnd,
				Metadata:    treasuryClient.FundSourceMetadata{BankMetadata: validBankMetadata},
			},
		},
		{
			name: "reject-empty-currency",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        gofakeit.Company(),
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.Zero,
				Currency:    shared.Currency{},
				Metadata:    treasuryClient.FundSourceMetadata{BankMetadata: validBankMetadata},
			},
		},
		{
			name: "reject-bank-type-with-cash-metadata",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        gofakeit.Company(),
				SourceType:  domain.FundSourceTypeBank,
				InitBalance: decimal.Zero,
				Currency:    vnd,
				Metadata:    treasuryClient.FundSourceMetadata{CashMetadata: validCashMetadata},
			},
		},
		{
			name: "reject-cash-type-with-bank-metadata",
			body: treasuryClient.RegisterFundSourceJSONRequestBody{
				Name:        gofakeit.Company(),
				SourceType:  domain.FundSourceTypeCash,
				InitBalance: decimal.Zero,
				Currency:    vnd,
				Metadata:    treasuryClient.FundSourceMetadata{BankMetadata: validBankMetadata},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.RegisterFundSourceWithResponse(ctx, tt.body)
			require.NoError(t, err)
			assertCorrelationID(t, resp.HTTPResponse)
			assert.Equal(t, 400, resp.StatusCode())
			assert.Nil(t, resp.JSON201)
		})
	}
}
