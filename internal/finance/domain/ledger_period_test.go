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

func TestNewLedgerPeriod(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validConfig := newValidPeriodConfig(t, 15, domain.PeriodFrequencyMonthly)

		tests := []struct {
			name        string
			accountUUID domain.LedgerAccountUUID
			config      domain.LedgerPeriodConfig
			wantErr     string
		}{
			{
				name:        "reject-zero-account-uuid",
				accountUUID: domain.LedgerAccountUUID{},
				config:      validConfig,
				wantErr:     "account uuid can't be empty",
			},
			{
				name:        "reject-zero-config",
				accountUUID: domain.LedgerAccountUUID{UUID: common.NewUUIDv7()},
				config:      domain.LedgerPeriodConfig{},
				wantErr:     "ledger period config can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewLedgerPeriod(tt.accountUUID, tt.config, decimal.Zero)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		accountUUID := domain.LedgerAccountUUID{UUID: common.NewUUIDv7()}
		config := newValidPeriodConfig(t, 15, domain.PeriodFrequencyMonthly)
		openBalance := decimal.NewFromInt(1_000_000)

		lp, err := domain.NewLedgerPeriod(accountUUID, config, openBalance)
		require.NoError(t, err)

		assert.False(t, lp.UUID().IsZero())
		assert.Equal(t, accountUUID, lp.AccountUUID())
		assert.Equal(t, config, lp.Config())
		assert.True(t, lp.OpenBalance().Equal(openBalance))
		assert.Equal(t, 15, lp.StartDate().Day())
		assert.True(t, lp.CloseDate().After(lp.StartDate()))
		assert.False(t, lp.IsClosed())
		assert.False(t, lp.IsReported())
		assert.True(t, lp.TotalDebit().IsZero())
		assert.True(t, lp.TotalCredit().IsZero())
	})
}

func TestNewNextLedgerPeriod(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validConfig := newValidPeriodConfig(t, 15, domain.PeriodFrequencyMonthly)
		closedPeriod := newClosedLedgerPeriod(t, decimal.NewFromInt(500_000))

		tests := []struct {
			name    string
			prev    *domain.LedgerPeriod
			config  domain.LedgerPeriodConfig
			wantErr string
		}{
			{
				name:    "reject-nil-previous",
				prev:    nil,
				config:  validConfig,
				wantErr: "previous ledger period can't be empty",
			},
			{
				name:    "reject-unclosed-previous",
				prev:    newOpenLedgerPeriod(t, decimal.Zero),
				config:  validConfig,
				wantErr: "current ledger period has not closed yet",
			},
			{
				name:    "reject-zero-config",
				prev:    closedPeriod,
				config:  domain.LedgerPeriodConfig{},
				wantErr: "ledger period config can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewNextLedgerPeriod(tt.prev, tt.config)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		openBalance := decimal.NewFromInt(2_000_000)
		prev := newClosedLedgerPeriod(t, openBalance)
		config := newValidPeriodConfig(t, 15, domain.PeriodFrequencyMonthly)

		next, err := domain.NewNextLedgerPeriod(prev, config)
		require.NoError(t, err)

		assert.False(t, next.UUID().IsZero())
		assert.NotEqual(t, prev.UUID(), next.UUID())
		assert.Equal(t, prev.AccountUUID(), next.AccountUUID())
		assert.True(t, next.OpenBalance().Equal(prev.CloseBalance()))
		assert.Equal(t, prev.CloseDate().AddDate(0, 0, 1), next.StartDate())
		assert.True(t, next.CloseDate().After(next.StartDate()))
		assert.False(t, next.IsClosed())
		assert.False(t, next.IsReported())
	})
}

func TestLedgerPeriod_Close(t *testing.T) {
	t.Run("reject-already-closed", func(t *testing.T) {
		t.Parallel()

		lp := newClosedLedgerPeriod(t, decimal.NewFromInt(500_000))

		err := lp.Close()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ledger period has already closed")
	})

	t.Run("closes-successfully", func(t *testing.T) {
		t.Parallel()

		openBalance := decimal.NewFromInt(1_000_000)
		lp := newOpenLedgerPeriodDay1(t, openBalance)
		inRange := lp.StartDate().AddDate(0, 0, 1)

		require.NoError(t, lp.PostJournalEntry(newJournalEntryWithDate(t, shared.EntryTypeDebit, inRange), false))
		require.NoError(t, lp.PostJournalEntry(newJournalEntryWithDate(t, shared.EntryTypeCredit, inRange), false))
		require.NoError(t, lp.PostJournalEntry(newJournalEntryWithDate(t, shared.EntryTypeCredit, inRange), false))

		err := lp.Close()
		require.NoError(t, err)

		// closeBalance = 1_000_000 + 200_000 - 100_000 = 1_100_000
		assert.True(t, lp.IsClosed())
		assert.True(t, lp.TotalDebit().Equal(decimal.NewFromInt(100_000)))
		assert.True(t, lp.TotalCredit().Equal(decimal.NewFromInt(200_000)))
		assert.True(t, lp.CloseBalance().Equal(decimal.NewFromInt(1_100_000)))
	})
}

func TestLedgerPeriod_PostJournalEntry(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(t *testing.T) (*domain.LedgerPeriod, *domain.JournalEntry, bool)
			wantErr string
		}{
			{
				name: "reject-nil-journal-entry",
				setup: func(t *testing.T) (*domain.LedgerPeriod, *domain.JournalEntry, bool) {
					return newOpenLedgerPeriodDay1(t, decimal.Zero), nil, false
				},
				wantErr: "journal entry can't be empty",
			},
			{
				name: "reject-closed-ledger-period",
				setup: func(t *testing.T) (*domain.LedgerPeriod, *domain.JournalEntry, bool) {
					lp := newClosedLedgerPeriod(t, decimal.Zero)
					je := newJournalEntryWithDate(t, shared.EntryTypeDebit, time.Now().UTC())
					return lp, je, false
				},
				wantErr: "ledger period has closed",
			},
			{
				name: "reject-transaction-date-before-period",
				setup: func(t *testing.T) (*domain.LedgerPeriod, *domain.JournalEntry, bool) {
					lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
					outOfRange := lp.StartDate().AddDate(0, 0, -1)
					je := newJournalEntryWithDate(t, shared.EntryTypeDebit, outOfRange)
					return lp, je, false
				},
				wantErr: "journal entry is not in ledger period",
			},
			{
				name: "reject-transaction-date-after-period",
				setup: func(t *testing.T) (*domain.LedgerPeriod, *domain.JournalEntry, bool) {
					lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
					outOfRange := lp.CloseDate().AddDate(0, 0, 1)
					je := newJournalEntryWithDate(t, shared.EntryTypeDebit, outOfRange)
					return lp, je, false
				},
				wantErr: "journal entry is not in ledger period",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				lp, je, forcePost := tt.setup(t)
				err := lp.PostJournalEntry(je, forcePost)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Run("posts-debit-entry-updates-total-debit", func(t *testing.T) {
			t.Parallel()

			lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
			amount := decimal.NewFromInt(100_000)
			je := newJournalEntryWithDate(t, shared.EntryTypeDebit, lp.StartDate().AddDate(0, 0, 1))

			err := lp.PostJournalEntry(je, false)
			require.NoError(t, err)

			assert.Len(t, lp.Entries(), 1)
			assert.True(t, lp.TotalDebit().Equal(amount))
			assert.True(t, lp.TotalCredit().IsZero())
		})

		t.Run("posts-credit-entry-updates-total-credit", func(t *testing.T) {
			t.Parallel()

			lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
			amount := decimal.NewFromInt(100_000)
			je := newJournalEntryWithDate(t, shared.EntryTypeCredit, lp.StartDate().AddDate(0, 0, 1))

			err := lp.PostJournalEntry(je, false)
			require.NoError(t, err)

			assert.Len(t, lp.Entries(), 1)
			assert.True(t, lp.TotalCredit().Equal(amount))
			assert.True(t, lp.TotalDebit().IsZero())
		})

		t.Run("force-post-allows-out-of-range-date", func(t *testing.T) {
			t.Parallel()

			lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
			outOfRange := lp.CloseDate().AddDate(0, 0, 1)
			je := newJournalEntryWithDate(t, shared.EntryTypeDebit, outOfRange)

			err := lp.PostJournalEntry(je, true)
			require.NoError(t, err)

			assert.Len(t, lp.Entries(), 1)
		})

		t.Run("accumulates-multiple-entries", func(t *testing.T) {
			t.Parallel()

			lp := newOpenLedgerPeriodDay1(t, decimal.Zero)
			inRange := lp.StartDate().AddDate(0, 0, 1)

			je1 := newJournalEntryWithDate(t, shared.EntryTypeDebit, inRange)
			je2 := newJournalEntryWithDate(t, shared.EntryTypeDebit, inRange)
			je3 := newJournalEntryWithDate(t, shared.EntryTypeCredit, inRange)

			require.NoError(t, lp.PostJournalEntry(je1, false))
			require.NoError(t, lp.PostJournalEntry(je2, false))
			require.NoError(t, lp.PostJournalEntry(je3, false))

			assert.Len(t, lp.Entries(), 3)
			assert.True(t, lp.TotalDebit().Equal(decimal.NewFromInt(200_000)))
			assert.True(t, lp.TotalCredit().Equal(decimal.NewFromInt(100_000)))
		})
	})
}

// helpers

func newValidPeriodConfig(t *testing.T, dayOfMonth int, frequency domain.PeriodFrequency) domain.LedgerPeriodConfig {
	t.Helper()
	cfg, err := domain.NewLedgerPeriodConfig(dayOfMonth, frequency)
	require.NoError(t, err)
	return cfg
}

func newOpenLedgerPeriodDay1(t *testing.T, openBalance decimal.Decimal) *domain.LedgerPeriod {
	t.Helper()
	accountUUID := domain.LedgerAccountUUID{UUID: common.NewUUIDv7()}
	config := newValidPeriodConfig(t, 1, domain.PeriodFrequencyMonthly)
	lp, err := domain.NewLedgerPeriod(accountUUID, config, openBalance)
	require.NoError(t, err)
	return lp
}

func newJournalEntryWithDate(t *testing.T, entryType shared.EntryType, transactionDate time.Time) *domain.JournalEntry {
	t.Helper()
	fpUUID := domain.FundProviderUUID{UUID: common.NewUUIDv7()}
	je, err := domain.NewJournalEntry(fpUUID, decimal.NewFromInt(100_000), entryType, transactionDate, nil)
	require.NoError(t, err)
	return je
}

func newOpenLedgerPeriod(t *testing.T, openBalance decimal.Decimal) *domain.LedgerPeriod {
	t.Helper()
	accountUUID := domain.LedgerAccountUUID{UUID: common.NewUUIDv7()}
	config := newValidPeriodConfig(t, 15, domain.PeriodFrequencyMonthly)
	lp, err := domain.NewLedgerPeriod(accountUUID, config, openBalance)
	require.NoError(t, err)
	return lp
}

func newClosedLedgerPeriod(t *testing.T, openBalance decimal.Decimal) *domain.LedgerPeriod {
	t.Helper()
	lp := newOpenLedgerPeriod(t, openBalance)
	err := lp.Close()
	require.NoError(t, err)
	return lp
}

func TestNewLedgerEntry(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		validAmount := decimal.NewFromInt(500_000)
		validPostingRef := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

		tests := []struct {
			name             string
			amount           decimal.Decimal
			entryType        shared.EntryType
			postingReference domain.JournalEntryUUID
			wantErr          string
		}{
			{
				name:             "reject-zero-amount",
				amount:           decimal.Zero,
				entryType:        shared.EntryTypeDebit,
				postingReference: validPostingRef,
				wantErr:          "amount can't be empty",
			},
			{
				name:             "reject-zero-entry-type",
				amount:           validAmount,
				entryType:        shared.EntryType{},
				postingReference: validPostingRef,
				wantErr:          "entry type can't be empty",
			},
			{
				name:             "reject-zero-posting-reference",
				amount:           validAmount,
				entryType:        shared.EntryTypeDebit,
				postingReference: domain.JournalEntryUUID{},
				wantErr:          "posting reference can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewLedgerEntry(tt.amount, tt.entryType, tt.postingReference)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		tests := []struct {
			name      string
			entryType shared.EntryType
		}{
			{
				name:      "debit-entry",
				entryType: shared.EntryTypeDebit,
			},
			{
				name:      "credit-entry",
				entryType: shared.EntryTypeCredit,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				amount := decimal.NewFromInt(500_000)
				postingRef := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

				entry, err := domain.NewLedgerEntry(amount, tt.entryType, postingRef)
				require.NoError(t, err)

				assert.True(t, entry.Amount().Equal(amount))
				assert.Equal(t, tt.entryType, entry.EntryType())
				assert.Equal(t, postingRef, entry.PostingReference())
				assert.False(t, entry.IsZero())
			})
		}
	})
}
