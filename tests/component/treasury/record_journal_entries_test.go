package treasury_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	treasuryClient "sumni-finance-backend/internal/treasury/api/http/client"
)

func TestRecordJournalEntries_DebitEntry_Returns201(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	fsUUID := registerBankSource(ctx, t, client, decimal.NewFromInt(1_000_000))

	recordJournalEntries(ctx, t, client, fsUUID, treasuryClient.RecordJournalEntriesJSONRequestBody{
		debitEntry(decimal.NewFromInt(500)),
	})
}

func TestRecordJournalEntries_MultipleEntries_Returns201(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	fsUUID := registerBankSource(ctx, t, client, decimal.NewFromInt(1_000_000))

	recordJournalEntries(ctx, t, client, fsUUID, treasuryClient.RecordJournalEntriesJSONRequestBody{
		debitEntry(decimal.NewFromInt(500_000)),
		creditEntry(decimal.NewFromInt(300_000)),
		debitEntry(decimal.NewFromInt(200_000)),
	})
}

func TestRecordJournalEntries_EmptyItems_Returns400(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	fsUUID := registerBankSource(ctx, t, client, decimal.Zero)

	resp, err := client.RecordJournalEntriesWithResponse(ctx, fsUUID, treasuryClient.RecordJournalEntriesJSONRequestBody{})
	require.NoError(t, err)
	assertCorrelationID(t, resp.HTTPResponse)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestRecordJournalEntries_DebitExceedsBalance_Returns400(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	fsUUID := registerBankSource(ctx, t, client, decimal.Zero)

	resp, err := client.RecordJournalEntriesWithResponse(ctx, fsUUID, treasuryClient.RecordJournalEntriesJSONRequestBody{
		debitEntry(decimal.NewFromInt(100)),
	})
	require.NoError(t, err)
	assertCorrelationID(t, resp.HTTPResponse)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestRecordJournalEntries_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newTestClient(t)
	fsUUID := registerBankSource(ctx, t, client, decimal.Zero)

	tests := []struct {
		name string
		body treasuryClient.RecordJournalEntriesJSONRequestBody
	}{
		{
			name: "reject-zero-amount",
			body: treasuryClient.RecordJournalEntriesJSONRequestBody{
				{
					Amount:          decimal.Zero,
					EntryType:       shared.EntryTypeDebit,
					TransactionDate: time.Now(),
				},
			},
		},
		{
			name: "reject-negative-amount",
			body: treasuryClient.RecordJournalEntriesJSONRequestBody{
				{
					Amount:          decimal.NewFromInt(-100_000),
					EntryType:       shared.EntryTypeDebit,
					TransactionDate: time.Now(),
				},
			},
		},
		{
			name: "reject-empty-entry-type",
			body: treasuryClient.RecordJournalEntriesJSONRequestBody{
				{
					Amount:          decimal.NewFromInt(100_000),
					EntryType:       shared.EntryType{},
					TransactionDate: time.Now(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.RecordJournalEntriesWithResponse(ctx, fsUUID, tt.body)
			require.NoError(t, err)
			assertCorrelationID(t, resp.HTTPResponse)
			assert.Equal(t, 400, resp.StatusCode())
		})
	}
}
