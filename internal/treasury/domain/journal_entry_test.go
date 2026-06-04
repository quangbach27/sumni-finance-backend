package domain_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury/domain"
)

func newFundSourceUUID() domain.FundSourceUUID {
	return domain.FundSourceUUID{UUID: common.NewUUIDv7()}
}

func newJournalEntry(t *testing.T, amount decimal.Decimal, entryType shared.EntryType) *domain.JournalEntry {
	t.Helper()
	je, err := domain.NewJournalEntry(
		amount,
		entryType,
		time.Now(),
		nil,
		nil,
		newFundSourceUUID(),
		decimal.Zero,
	)
	require.NoError(t, err)
	return je
}

// NewJournalEntry

func TestNewJournalEntry_Success(t *testing.T) {
	t.Parallel()

	amount := decimal.NewFromInt(100_000)
	transactionNo := "TXN-001"
	transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	fsUUID := newFundSourceUUID()
	balanceBefore := decimal.NewFromInt(500_000)

	je, err := domain.NewJournalEntry(
		amount,
		shared.EntryTypeDebit,
		transactionDate,
		&transactionNo,
		nil,
		fsUUID,
		balanceBefore,
	)

	require.NoError(t, err)
	assert.False(t, je.UUID().IsZero())
	assert.Equal(t, "100000", je.Amount().String())
	assert.Equal(t, shared.EntryTypeDebit, je.EntryType())
	assert.Equal(t, transactionDate, je.TransactionDate())
	assert.Equal(t, &transactionNo, je.TransactionNo())
	assert.Equal(t, domain.JournalEntryStatusRecorded, je.Status())
	assert.Equal(t, fsUUID, je.FundSourceUUID())
	assert.Equal(t, "500000", je.BalanceBefore().String())
	assert.Equal(t, "0", je.BalanceAfter().String())
	assert.Nil(t, je.Description())
	assert.Nil(t, je.VoidedBy())
	assert.Nil(t, je.VoidedAt())
	assert.Nil(t, je.VoidedReason())
	assert.Nil(t, je.ReverseEntryUUID())
}

func TestNewJournalEntry_WithDescription(t *testing.T) {
	t.Parallel()

	desc := "payment for invoice #42"
	je, err := domain.NewJournalEntry(
		decimal.NewFromInt(1000),
		shared.EntryTypeDebit,
		time.Now(),
		nil,
		&desc,
		newFundSourceUUID(),
		decimal.Zero,
	)

	require.NoError(t, err)
	require.NotNil(t, je.Description())
	assert.Equal(t, "payment for invoice #42", *je.Description())
}

func TestNewJournalEntry_NilTransactionNo(t *testing.T) {
	t.Parallel()

	je, err := domain.NewJournalEntry(
		decimal.NewFromInt(1000),
		shared.EntryTypeCredit,
		time.Now(),
		nil,
		nil,
		newFundSourceUUID(),
		decimal.Zero,
	)

	require.NoError(t, err)
	assert.Nil(t, je.TransactionNo())
}

func TestNewJournalEntry_ValidationErrors(t *testing.T) {
	t.Parallel()

	validAmount := decimal.NewFromInt(1000)
	validDate := time.Now()
	validFSUUID := newFundSourceUUID()

	tests := []struct {
		name            string
		amount          decimal.Decimal
		entryType       shared.EntryType
		transactionDate time.Time
		fundSourceUUID  domain.FundSourceUUID
		wantErrSlugs    []string
	}{
		{
			name:            "reject-zero-amount",
			amount:          decimal.Zero,
			entryType:       shared.EntryTypeDebit,
			transactionDate: validDate,
			fundSourceUUID:  validFSUUID,
			wantErrSlugs:    []string{"invalid-amount"},
		},
		{
			name:            "reject-negative-amount",
			amount:          decimal.NewFromInt(-1),
			entryType:       shared.EntryTypeDebit,
			transactionDate: validDate,
			fundSourceUUID:  validFSUUID,
			wantErrSlugs:    []string{"invalid-amount"},
		},
		{
			name:            "reject-empty-entry-type",
			amount:          validAmount,
			entryType:       shared.EntryType{},
			transactionDate: validDate,
			fundSourceUUID:  validFSUUID,
			wantErrSlugs:    []string{"empty-entry-type"},
		},
		{
			name:            "reject-zero-transaction-date",
			amount:          validAmount,
			entryType:       shared.EntryTypeDebit,
			transactionDate: time.Time{},
			fundSourceUUID:  validFSUUID,
			wantErrSlugs:    []string{"empty-transaction-date"},
		},
		{
			name:            "reject-zero-fund-source-uuid",
			amount:          validAmount,
			entryType:       shared.EntryTypeDebit,
			transactionDate: validDate,
			fundSourceUUID:  domain.FundSourceUUID{},
			wantErrSlugs:    []string{"empty-fund-source-uuid"},
		},
		{
			name:            "reject-multiple-validation-errors",
			amount:          decimal.Zero,
			entryType:       shared.EntryType{},
			transactionDate: time.Time{},
			fundSourceUUID:  domain.FundSourceUUID{},
			wantErrSlugs:    []string{"invalid-amount", "empty-entry-type", "empty-transaction-date", "empty-fund-source-uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewJournalEntry(
				tt.amount,
				tt.entryType,
				tt.transactionDate,
				nil,
				nil,
				tt.fundSourceUUID,
				decimal.Zero,
			)

			require.Error(t, err)
			expectedDetails := make([]common.ErrorDetails, len(tt.wantErrSlugs))
			for i, slug := range tt.wantErrSlugs {
				expectedDetails[i] = common.ErrorDetails{ErrorSlug: slug}
			}
			testutils.AssertErrorDetails(t, err, expectedDetails)
		})
	}
}

// IsDebitSide

func TestJournalEntry_IsDebitSide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entryType shared.EntryType
		want      bool
	}{
		{name: "debit-returns-true", entryType: shared.EntryTypeDebit, want: true},
		{name: "credit-returns-false", entryType: shared.EntryTypeCredit, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			je := newJournalEntry(t, decimal.NewFromInt(1000), tt.entryType)
			assert.Equal(t, tt.want, je.IsDebitSide())
		})
	}
}

// SetBalanceBefore / SetBalanceAfter

func TestJournalEntry_SetBalanceBefore(t *testing.T) {
	t.Parallel()

	je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
	je.SetBalanceBefore(decimal.NewFromInt(200_000))

	assert.Equal(t, "200000", je.BalanceBefore().String())
}

func TestJournalEntry_SetBalanceAfter(t *testing.T) {
	t.Parallel()

	je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
	je.SetBalanceAfter(decimal.NewFromInt(201_000))

	assert.Equal(t, "201000", je.BalanceAfter().String())
}

// MarkAsPosted

func TestJournalEntry_MarkAsPosted_Success(t *testing.T) {
	t.Parallel()

	je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)

	err := je.MarkAsPosted("user-1")

	require.NoError(t, err)
	assert.Equal(t, domain.JournalEntryStatusPosted, je.Status())
}

func TestJournalEntry_MarkAsPosted_ValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("reject-already-posted", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
		require.NoError(t, je.MarkAsPosted("user-1"))

		err := je.MarkAsPosted("user-1")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "RECORDED")
	})

	t.Run("reject-voided", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
		_, err := je.Void("user-1", time.Now().Add(time.Hour), "test reason")
		require.NoError(t, err)

		err = je.MarkAsPosted("user-1")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "RECORDED")
	})
}

// Void

func TestJournalEntry_Void_Success(t *testing.T) {
	t.Parallel()

	je := newJournalEntry(t, decimal.NewFromInt(50_000), shared.EntryTypeDebit)
	voidedAt := time.Now().Add(time.Hour)

	reverse, err := je.Void("user-2", voidedAt, "duplicate entry")

	require.NoError(t, err)
	require.NotNil(t, reverse)

	assert.Equal(t, domain.JournalEntryStatusVoided, je.Status())
	assert.Equal(t, "user-2", *je.VoidedBy())
	assert.Equal(t, voidedAt, *je.VoidedAt())
	assert.Equal(t, "duplicate entry", *je.VoidedReason())
	assert.NotNil(t, je.ReverseEntryUUID())

	assert.Equal(t, "50000", reverse.Amount().String())
	assert.Equal(t, shared.EntryTypeDebit, reverse.EntryType())
	assert.Equal(t, je.FundSourceUUID(), reverse.FundSourceUUID())
}

func TestJournalEntry_Void_AlreadyVoided(t *testing.T) {
	t.Parallel()

	je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeCredit)
	_, err := je.Void("user-1", time.Now().Add(time.Hour), "first void")
	require.NoError(t, err)

	_, err = je.Void("user-1", time.Now().Add(2*time.Hour), "second void")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already voided")
}

func TestJournalEntry_Void_ValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("reject-empty-voided-by", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)

		_, err := je.Void("", time.Now().Add(time.Hour), "some reason")

		require.Error(t, err)
		testutils.AssertErrorDetails(t, err, []common.ErrorDetails{
			{ErrorSlug: "empty-voided-by"},
		})
	})

	t.Run("reject-voided-at-before-created-at", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
		pastTime := time.Now().Add(-time.Hour)

		_, err := je.Void("user-1", pastTime, "some reason")

		require.Error(t, err)
		testutils.AssertErrorDetails(t, err, []common.ErrorDetails{
			{ErrorSlug: "invalid-voided-at"},
		})
	})

	t.Run("reject-empty-voided-reason", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)

		_, err := je.Void("user-1", time.Now().Add(time.Hour), "")

		require.Error(t, err)
		testutils.AssertErrorDetails(t, err, []common.ErrorDetails{
			{ErrorSlug: "empty-voided-reason"},
		})
	})

	t.Run("reject-multiple-validation-errors", func(t *testing.T) {
		t.Parallel()

		je := newJournalEntry(t, decimal.NewFromInt(1000), shared.EntryTypeDebit)
		pastTime := time.Now().Add(-time.Hour)

		_, err := je.Void("", pastTime, "")

		require.Error(t, err)
		testutils.AssertErrorDetails(t, err, []common.ErrorDetails{
			{ErrorSlug: "empty-voided-by"},
			{ErrorSlug: "invalid-voided-at"},
			{ErrorSlug: "empty-voided-reason"},
		})
	})
}
