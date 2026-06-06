//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	repoDb "sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/app/query"
	"sumni-finance-backend/internal/treasury/domain"
)

var jeViewCmpOpts = []cmp.Option{
	cmp.AllowUnexported(
		common.Enum[shared.EntryTypeValues]{},
		common.Enum[domain.JournalEntryStatusValues]{},
	),
	cmpopts.EquateApproxTime(time.Second),
}

func TestListJournalEntries_EmptyResult(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(100_000))

	result, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           1,
		PageSize:       20,
	})

	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, 0, result.TotalItems)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
}

func TestListJournalEntries_FieldMapping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))

	txNo := "TXN-001"
	desc := "salary deposit"
	transactionDate := time.Now().UTC().Truncate(time.Millisecond)

	entry := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeCredit, transactionDate, &txNo, &desc)

	result, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           1,
		PageSize:       20,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	want := expectedJournalEntryView(entry)
	if diff := cmp.Diff(want, result.Items[0], jeViewCmpOpts...); diff != "" {
		t.Errorf("journal entry view mismatch (-want +got):\n%s", diff)
	}
}

func TestListJournalEntries_DoesNotIncludeOtherFundSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs1 := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))
	fs2 := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))

	entry1 := mustRecordEntry(t, ctx, repo, fs1.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, time.Now(), nil, nil)
	mustRecordEntry(t, ctx, repo, fs2.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, time.Now(), nil, nil)

	result, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs1.UUID(),
		Page:           1,
		PageSize:       20,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, 1, result.TotalItems)
	assert.Equal(t, entry1.UUID(), result.Items[0].UUID)
}

func TestListJournalEntries_Pagination(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(1_000_000))

	now := time.Now().UTC()
	e1 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-2*time.Hour), nil, nil)
	e2 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-1*time.Hour), nil, nil)
	e3 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now, nil, nil)

	page1, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           1,
		PageSize:       2,
	})
	require.NoError(t, err)
	assert.Len(t, page1.Items, 2)
	assert.Equal(t, 3, page1.TotalItems)
	assert.Equal(t, 1, page1.Page)
	assert.Equal(t, 2, page1.PageSize)
	// ordered transaction_date DESC: e3 first, then e2
	assert.Equal(t, e3.UUID(), page1.Items[0].UUID)
	assert.Equal(t, e2.UUID(), page1.Items[1].UUID)

	page2, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           2,
		PageSize:       2,
	})
	require.NoError(t, err)
	assert.Len(t, page2.Items, 1)
	assert.Equal(t, 3, page2.TotalItems)
	assert.Equal(t, e1.UUID(), page2.Items[0].UUID)
}

func TestListJournalEntries_DateFromFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(1_000_000))

	now := time.Now().UTC()
	mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-48*time.Hour), nil, nil)
	e2 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-24*time.Hour), nil, nil)
	e3 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now, nil, nil)

	dateFrom := now.Add(-36 * time.Hour)
	result, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           1,
		PageSize:       20,
		DateFrom:       &dateFrom,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.TotalItems)
	assert.Equal(t, e3.UUID(), result.Items[0].UUID)
	assert.Equal(t, e2.UUID(), result.Items[1].UUID)
}

func TestListJournalEntries_DateToFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewJournalEntryReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(1_000_000))

	now := time.Now().UTC()
	e1 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-48*time.Hour), nil, nil)
	e2 := mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now.Add(-24*time.Hour), nil, nil)
	mustRecordEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000),
		shared.EntryTypeDebit, now, nil, nil)

	dateTo := now.Add(-12 * time.Hour)
	result, err := readModel.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: fs.UUID(),
		Page:           1,
		PageSize:       20,
		DateTo:         &dateTo,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.TotalItems)
	assert.Equal(t, e2.UUID(), result.Items[0].UUID)
	assert.Equal(t, e1.UUID(), result.Items[1].UUID)
}

func mustRecordEntry(
	t *testing.T,
	ctx context.Context,
	repo *repoDb.FundSourceRepo,
	fsUUID domain.FundSourceUUID,
	amount decimal.Decimal,
	entryType shared.EntryType,
	transactionDate time.Time,
	transactionNo *string,
	description *string,
) *domain.JournalEntry {
	t.Helper()

	var captured *domain.JournalEntry

	err := repo.RecordJournalEntries(ctx, fsUUID, func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		money, err := shared.NewMoney(amount, vnd)
		if err != nil {
			return nil, err
		}

		balanceBefore := fs.Balance().Amount()

		if entryType == shared.EntryTypeDebit {
			if err = fs.Withdraw(money); err != nil {
				return nil, err
			}
		} else {
			if err = fs.TopUp(money); err != nil {
				return nil, err
			}
		}

		entry, err := domain.NewJournalEntry(amount, entryType, transactionDate, transactionNo, description, fsUUID, balanceBefore)
		if err != nil {
			return nil, err
		}
		if err = entry.SetBalanceAfter(fs.Balance().Amount()); err != nil {
			return nil, err
		}

		captured = entry
		return []*domain.JournalEntry{entry}, nil
	})

	require.NoError(t, err)
	return captured
}

func expectedJournalEntryView(je *domain.JournalEntry) query.JournalEntryView {
	return query.JournalEntryView{
		UUID:             je.UUID(),
		FundSourceUUID:   je.FundSourceUUID(),
		Amount:           je.Amount(),
		EntryType:        je.EntryType(),
		TransactionDate:  je.TransactionDate(),
		TransactionNo:    je.TransactionNo(),
		Description:      je.Description(),
		Status:           je.Status(),
		BalanceBefore:    je.BalanceBefore(),
		BalanceAfter:     je.BalanceAfter(),
		ReverseEntryUUID: je.ReverseEntryUUID(),
		VoidedBy:         je.VoidedBy(),
		VoidedAt:         je.VoidedAt(),
		VoidedReason:     je.VoidedReason(),
		CreatedAt:        je.CreatedAt(),
		CreatedBy:        je.CreatedBy(),
		UpdatedAt:        je.UpdatedAt(),
		UpdatedBy:        je.UpdatedBy(),
	}
}
