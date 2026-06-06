package treasury_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	treasuryClient "sumni-finance-backend/internal/treasury/api/http/client"
	"sumni-finance-backend/internal/treasury/domain"
)

func registerBankSource(
	ctx context.Context,
	t *testing.T,
	client testClient,
	initBalance decimal.Decimal,
) domain.FundSourceUUID {
	t.Helper()

	vnd := shared.MustNewCurrency("VND")

	resp, err := client.RegisterFundSourceWithResponse(
		ctx,
		treasuryClient.RegisterFundSourceJSONRequestBody{
			Name:        gofakeit.Company(),
			SourceType:  domain.FundSourceTypeBank,
			InitBalance: initBalance,
			Currency:    vnd,
			Metadata: treasuryClient.FundSourceMetadata{
				BankMetadata: &treasuryClient.FundSourceBankMetadata{
					BankCode:      "VCB",
					AccountNumber: gofakeit.Numerify("##########"),
					AccountOwner:  gofakeit.Name(),
				},
			},
		},
	)

	require.NoError(t, err)
	assertCorrelationID(t, resp.HTTPResponse)
	require.NotNil(t, resp.JSON201, "expected 201 response, got: %s", resp.Status())

	return resp.JSON201.FundSourceUuid
}

func recordJournalEntries(
	ctx context.Context,
	t *testing.T,
	client testClient,
	fsUUID domain.FundSourceUUID,
	items treasuryClient.RecordJournalEntriesJSONRequestBody,
) {
	t.Helper()

	resp, err := client.RecordJournalEntriesWithResponse(ctx, fsUUID, items)

	require.NoError(t, err)
	assertCorrelationID(t, resp.HTTPResponse)
	require.Equal(t, 201, resp.StatusCode(), "expected 201 response, got: %s", resp.Status())
}

func debitEntry(amount decimal.Decimal) treasuryClient.JournalEntryItem {
	return treasuryClient.JournalEntryItem{
		Amount:          amount,
		EntryType:       shared.EntryTypeDebit,
		TransactionDate: time.Now(),
	}
}

func creditEntry(amount decimal.Decimal) treasuryClient.JournalEntryItem {
	return treasuryClient.JournalEntryItem{
		Amount:          amount,
		EntryType:       shared.EntryTypeCredit,
		TransactionDate: time.Now(),
	}
}
