package domain_test

import (
	"testing"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJournalEntry(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validFPUUID := domain.FundProviderUUID{UUID: common.NewUUIDv7()}
		validAmount := decimal.NewFromInt(100_000)
		validEntryType := shared.EntryTypeDebit
		validDate := time.Now().UTC()

		tests := []struct {
			name             string
			fundProviderUUID domain.FundProviderUUID
			amount           decimal.Decimal
			entryType        shared.EntryType
			transactionDate  time.Time
		}{
			{
				name:             "reject-zero-fund-provider-uuid",
				fundProviderUUID: domain.FundProviderUUID{},
				amount:           validAmount,
				entryType:        validEntryType,
				transactionDate:  validDate,
			},
			{
				name:             "reject-zero-amount",
				fundProviderUUID: validFPUUID,
				amount:           decimal.Zero,
				entryType:        validEntryType,
				transactionDate:  validDate,
			},
			{
				name:             "reject-zero-entry-type",
				fundProviderUUID: validFPUUID,
				amount:           validAmount,
				entryType:        shared.EntryType{},
				transactionDate:  validDate,
			},
			{
				name:             "reject-zero-transaction-date",
				fundProviderUUID: validFPUUID,
				amount:           validAmount,
				entryType:        validEntryType,
				transactionDate:  time.Time{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewJournalEntry(tt.fundProviderUUID, tt.amount, tt.entryType, tt.transactionDate, nil)
				require.Error(t, err)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		fpUUID := domain.FundProviderUUID{UUID: common.NewUUIDv7()}
		transactionNo := "TXN-001"
		txDate := time.Now().UTC()
		amount := decimal.NewFromInt(100_000)

		je, err := domain.NewJournalEntry(fpUUID, amount, shared.EntryTypeDebit, txDate, &transactionNo)
		require.NoError(t, err)

		assert.False(t, je.UUID().IsZero())
		assert.Equal(t, fpUUID, je.FundProviderUUID())
		assert.True(t, je.Amount().Equal(amount))
		assert.Equal(t, shared.EntryTypeDebit, je.EntryType())
		assert.Equal(t, txDate, je.TransactionDate())
		assert.Equal(t, &transactionNo, je.TransactionNo())
		assert.True(t, je.IsDraft())
		assert.False(t, je.IsPosted())
		assert.False(t, je.IsVoided())
		assert.Nil(t, je.ReverseEntry())
		assert.Nil(t, je.VoidReason())
		assert.Nil(t, je.VoidDate())
		assert.Nil(t, je.VoidBy())
	})
}

func TestNewReverseJournalEntry(t *testing.T) {
	t.Run("reject-nil-previous-entry", func(t *testing.T) {
		t.Parallel()

		_, err := domain.NewReverseJournalEntry(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "previous journal entry can't be empty")
	})

	t.Run("reject-draft-previous-entry", func(t *testing.T) {
		t.Parallel()

		prevEntry := newValidJournalEntry(t, shared.EntryTypeDebit)

		_, err := domain.NewReverseJournalEntry(prevEntry)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "previous entry should be posted before create reverse journal entry")
	})

	t.Run("success-from-posted-entry", func(t *testing.T) {
		t.Parallel()

		prevEntry := newPostedJournalEntry(t, shared.EntryTypeDebit)

		reverse, err := domain.NewReverseJournalEntry(prevEntry)
		require.NoError(t, err)

		assert.False(t, reverse.UUID().IsZero())
		assert.True(t, reverse.IsDraft())
	})

	t.Run("success-from-voided-entry", func(t *testing.T) {
		t.Parallel()

		prevEntry := newPostedJournalEntry(t, shared.EntryTypeDebit)
		reverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}
		err := prevEntry.Void(reverseUUID, "duplicate entry", "admin")
		require.NoError(t, err)

		reverse, err := domain.NewReverseJournalEntry(prevEntry)
		require.NoError(t, err)

		assert.False(t, reverse.UUID().IsZero())
		assert.True(t, reverse.IsDraft())
	})
}

func TestJournalEntry_Post(t *testing.T) {
	t.Run("posts-draft-entry-successfully", func(t *testing.T) {
		t.Parallel()

		je := newValidJournalEntry(t, shared.EntryTypeDebit)

		err := je.Post()
		require.NoError(t, err)

		assert.True(t, je.IsPosted())
		assert.False(t, je.IsDraft())
	})

	t.Run("reject-post-on-already-posted-entry", func(t *testing.T) {
		t.Parallel()

		je := newPostedJournalEntry(t, shared.EntryTypeDebit)

		err := je.Post()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "journal entry has already posted")
	})

	t.Run("reject-post-on-voided-entry", func(t *testing.T) {
		t.Parallel()

		je := newPostedJournalEntry(t, shared.EntryTypeDebit)
		reverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}
		err := je.Void(reverseUUID, "test void", "admin")
		require.NoError(t, err)

		err = je.Post()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "journal entry has already posted")
	})
}

func TestJournalEntry_Void(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validReverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

		tests := []struct {
			name         string
			reverseEntry domain.JournalEntryUUID
			voidReason   string
			voidBy       string
			wantErr      string
		}{
			{
				name:         "reject-zero-reverse-entry",
				reverseEntry: domain.JournalEntryUUID{},
				voidReason:   "duplicate entry",
				voidBy:       "admin",
				wantErr:      "reverse entry can't be empty",
			},
			{
				name:         "reject-empty-void-reason",
				reverseEntry: validReverseUUID,
				voidReason:   "",
				voidBy:       "admin",
				wantErr:      "void reason can't be empty",
			},
			{
				name:         "reject-empty-void-by",
				reverseEntry: validReverseUUID,
				voidReason:   "duplicate entry",
				voidBy:       "",
				wantErr:      "void by can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				je := newPostedJournalEntry(t, shared.EntryTypeDebit)
				err := je.Void(tt.reverseEntry, tt.voidReason, tt.voidBy)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("reject-void-on-draft-entry", func(t *testing.T) {
		t.Parallel()

		je := newValidJournalEntry(t, shared.EntryTypeDebit)
		reverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

		err := je.Void(reverseUUID, "duplicate entry", "admin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft entry can't be voided")
	})

	t.Run("reject-void-on-already-voided-entry", func(t *testing.T) {
		t.Parallel()

		je := newPostedJournalEntry(t, shared.EntryTypeDebit)
		reverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

		err := je.Void(reverseUUID, "duplicate entry", "admin")
		require.NoError(t, err)

		err = je.Void(reverseUUID, "duplicate entry", "admin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "journal entry has already voided")
	})

	t.Run("voids-posted-entry-successfully", func(t *testing.T) {
		t.Parallel()

		je := newPostedJournalEntry(t, shared.EntryTypeDebit)
		reverseUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

		err := je.Void(reverseUUID, "duplicate entry", "admin")
		require.NoError(t, err)

		assert.True(t, je.IsVoided())
		assert.False(t, je.IsPosted())
		assert.False(t, je.IsDraft())
		require.NotNil(t, je.ReverseEntry())
		assert.Equal(t, reverseUUID, *je.ReverseEntry())
		require.NotNil(t, je.VoidReason())
		assert.Equal(t, "duplicate entry", *je.VoidReason())
		require.NotNil(t, je.VoidBy())
		assert.Equal(t, "admin", *je.VoidBy())
		assert.NotNil(t, je.VoidDate())
	})
}

func newValidJournalEntry(t *testing.T, entryType shared.EntryType) *domain.JournalEntry {
	t.Helper()
	fpUUID := domain.FundProviderUUID{UUID: common.NewUUIDv7()}
	je, err := domain.NewJournalEntry(fpUUID, decimal.NewFromInt(100_000), entryType, time.Now().UTC(), nil)
	require.NoError(t, err)
	return je
}

func newPostedJournalEntry(t *testing.T, entryType shared.EntryType) *domain.JournalEntry {
	t.Helper()
	je := newValidJournalEntry(t, entryType)
	err := je.Post()
	require.NoError(t, err)
	return je
}
