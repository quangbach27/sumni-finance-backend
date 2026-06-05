//go:build integration

package db_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury/adapters/bank/lookup"
	repoDb "sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"
)

var fundSourceFactory = domain.NewFundSourceFactory(lookup.NewStubClient())

var vnd = shared.MustNewCurrency("VND")

var cmpOpts = []cmp.Option{
	cmp.AllowUnexported(
		common.Enum[domain.FundSourceTypeValues]{},
		common.Enum[shared.CurrencyType]{},
		domain.BankInfo{},
	),
	cmpopts.EquateApproxTime(time.Second),
}

func TestSaveFundSource_NilError(t *testing.T) {
	t.Parallel()

	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	err := repo.SaveFundSource(context.Background(), nil)
	require.Error(t, err)
}

func TestSaveFundSource_BankFundSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	balance, err := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
	require.NoError(t, err)

	metadata, err := fundSourceFactory.NewBankMetadata(ctx, "VCB", gofakeit.Numerify("##########"), gofakeit.Name())
	require.NoError(t, err)

	fs, err := fundSourceFactory.NewFundSource("Techcombank-SRB", domain.FundSourceTypeBank, balance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, repo.SaveFundSource(ctx, fs))

	queries := dbmodels.New(db)
	actualRow, err := queries.GetFundSourceByUUID(ctx, fs.UUID())
	require.NoError(t, err)

	if diff := cmp.Diff(fundSourceToExpectedRow(fs), actualRow, cmpOpts...); diff != "" {
		t.Errorf("saved row mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveFundSource_CashFundSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	balance, err := shared.NewMoney(decimal.NewFromInt(200_000), vnd)
	require.NoError(t, err)

	metadata, err := fundSourceFactory.NewCashMetadata("John Doe")
	require.NoError(t, err)

	fs, err := fundSourceFactory.NewFundSource("Cash Wallet", domain.FundSourceTypeCash, balance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, repo.SaveFundSource(ctx, fs))

	queries := dbmodels.New(db)
	actualRow, err := queries.GetFundSourceByUUID(ctx, fs.UUID())
	require.NoError(t, err)

	if diff := cmp.Diff(fundSourceToExpectedRow(fs), actualRow, cmpOpts...); diff != "" {
		t.Errorf("saved row mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveFundSource_DuplicateBankAccountReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	balance, err := shared.NewMoney(decimal.Zero, vnd)
	require.NoError(t, err)

	accountNumber := testutils.RandomString(10)
	accountOwner := testutils.RandomString(12)

	metadata, err := fundSourceFactory.NewBankMetadata(ctx, "VCB", accountNumber, accountOwner)
	require.NoError(t, err)

	first, err := fundSourceFactory.NewFundSource("First Account", domain.FundSourceTypeBank, balance, vnd, metadata)
	require.NoError(t, err)
	require.NoError(t, repo.SaveFundSource(ctx, first))

	second, err := fundSourceFactory.NewFundSource("Second Account", domain.FundSourceTypeBank, balance, vnd, metadata)
	require.NoError(t, err)

	err = repo.SaveFundSource(ctx, second)
	require.Error(t, err, "expected error when saving fund source with duplicate bank account number and owner")
}

func TestRecordJournalEntries_FundSourceNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	nonExistentUUID := domain.FundSourceUUID{UUID: common.NewUUIDv7()}

	err := repo.RecordJournalEntries(ctx, nonExistentUUID, func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		return nil, nil
	})

	require.Error(t, err)
}

func TestRecordJournalEntries_RecordFnError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(100_000))

	err := repo.RecordJournalEntries(ctx, fs.UUID(), func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		return nil, errors.New("recordFn error")
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "recordFn error")
}

func TestRecordJournalEntries_DebitEntry_Wtihdraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	initialBalance := decimal.NewFromInt(500_000)
	entryAmount := decimal.NewFromInt(200_000)

	fs := mustCreateCashFundSource(t, ctx, repo, initialBalance)

	err := repo.RecordJournalEntries(ctx, fs.UUID(), func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		money, err := shared.NewMoney(entryAmount, vnd)
		if err != nil {
			return nil, err
		}

		balanceBefore := fs.Balance().Amount()
		if err = fs.Withdraw(money); err != nil {
			return nil, err
		}

		entry, err := domain.NewJournalEntry(entryAmount, shared.EntryTypeDebit, time.Now(), nil, nil, fs.UUID(), balanceBefore)
		if err != nil {
			return nil, err
		}
		err = entry.SetBalanceAfter(fs.Balance().Amount())
		if err != nil {
			return nil, err
		}

		return []*domain.JournalEntry{entry}, nil
	})

	require.NoError(t, err)
}

func TestRecordJournalEntries_CreditEntry_TopUp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	initialBalance := decimal.NewFromInt(500_000)
	entryAmount := decimal.NewFromInt(150_000)

	fs := mustCreateCashFundSource(t, ctx, repo, initialBalance)

	err := repo.RecordJournalEntries(ctx, fs.UUID(), func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		money, err := shared.NewMoney(entryAmount, vnd)
		if err != nil {
			return nil, err
		}

		balanceBefore := fs.Balance().Amount()
		if err = fs.TopUp(money); err != nil {
			return nil, err
		}

		entry, err := domain.NewJournalEntry(entryAmount, shared.EntryTypeCredit, time.Now(), nil, nil, fs.UUID(), balanceBefore)
		if err != nil {
			return nil, err
		}
		err = entry.SetBalanceAfter(fs.Balance().Amount())
		if err != nil {
			return nil, err
		}

		return []*domain.JournalEntry{entry}, nil
	})

	require.NoError(t, err)
}

func TestRecordJournalEntries_MultipleEntries_Concurrent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	initialBalance := decimal.NewFromInt(1_000_000)

	entryAmounts := []struct {
		amount    decimal.Decimal
		entryType shared.EntryType
	}{
		{
			amount:    decimal.NewFromInt(300_000),
			entryType: shared.EntryTypeDebit,
		},
		{
			amount:    decimal.NewFromInt(300_000),
			entryType: shared.EntryTypeDebit,
		},
		{
			amount:    decimal.NewFromInt(300_000),
			entryType: shared.EntryTypeCredit,
		},
		{
			amount:    decimal.NewFromInt(300_000),
			entryType: shared.EntryTypeCredit,
		},
	}

	fs := mustCreateCashFundSource(t, ctx, repo, initialBalance)

	var wg sync.WaitGroup

	for _, entryAmount := range entryAmounts {
		wg.Go(func() {
			err := repo.RecordJournalEntries(ctx, fs.UUID(), func(fundSource *domain.FundSource) ([]*domain.JournalEntry, error) {
				money, err := shared.NewMoney(entryAmount.amount, vnd)
				if err != nil {
					return nil, err
				}

				balanceBefore := fundSource.Balance().Amount()

				if entryAmount.entryType == shared.EntryTypeDebit {
					err = fundSource.Withdraw(money)
				} else {
					err = fundSource.TopUp(money)
				}
				if err != nil {
					return nil, err
				}

				entry, err := domain.NewJournalEntry(entryAmount.amount, entryAmount.entryType, time.Now(), nil, nil, fundSource.UUID(), balanceBefore)
				if err != nil {
					return nil, err
				}

				entry.SetBalanceAfter(fundSource.Balance().Amount())

				return []*domain.JournalEntry{entry}, nil
			})
			assert.NoError(t, err)
		})
	}

	wg.Wait()

	expectedBalance := initialBalance

	updatedRow, err := dbmodels.New(db).GetFundSourceByUUID(ctx, fs.UUID())
	require.NoError(t, err)
	require.True(t, expectedBalance.Equal(updatedRow.Balance))
}

func TestVoidJournalEntry_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))
	entryUUID := mustRecordDebitEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000))

	voidedAt := time.Now().Add(time.Second)
	reverseEntry, err := repo.VoidJournalEntry(
		ctx,
		fs.UUID(),
		entryUUID,
		func(fs *domain.FundSource, je *domain.JournalEntry) (*domain.JournalEntry, error) {
			reverse, err := je.Void("actor-1", voidedAt, "duplicate entry")
			if err != nil {
				return nil, err
			}

			money, err := shared.NewMoney(je.Amount(), fs.Currency())
			if err != nil {
				return nil, err
			}

			balanceBefore := fs.Balance().Amount()
			if reverse.IsDebitSide() {
				if err = fs.Withdraw(money); err != nil {
					return nil, err
				}
			} else {
				if err = fs.TopUp(money); err != nil {
					return nil, err
				}
			}
			reverse.SetBalanceBefore(balanceBefore)
			reverse.SetBalanceAfter(fs.Balance().Amount())

			return reverse, nil
		},
	)

	require.NoError(t, err)
	require.NotNil(t, reverseEntry)
	assert.False(t, reverseEntry.UUID().IsZero())
	assert.True(t, reverseEntry.Status().Equal(domain.JournalEntryStatusReverse.Enum))

	queries := dbmodels.New(db)

	original, err := queries.GetJournalEntryByUID(ctx, entryUUID)
	require.NoError(t, err)
	assert.Equal(t, domain.JournalEntryStatusVoided, original.Status)
	assert.Equal(t, "actor-1", *original.VoidedBy)
	assert.Equal(t, "duplicate entry", *original.VoidedReason)
	assert.NotNil(t, original.ReverseEntryUuid)

	reverse, err := queries.GetJournalEntryByUID(ctx, reverseEntry.UUID())
	require.NoError(t, err)
	assert.Equal(t, domain.JournalEntryStatusReverse, reverse.Status)
	assert.Equal(t, fs.UUID(), reverse.FundSourceUuid)
}

func TestVoidJournalEntry_FundSourceNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	nonExistentFSUUID := domain.FundSourceUUID{UUID: common.NewUUIDv7()}
	nonExistentJEUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

	_, err := repo.VoidJournalEntry(
		ctx,
		nonExistentFSUUID,
		nonExistentJEUUID,
		func(_ *domain.FundSource, je *domain.JournalEntry) (*domain.JournalEntry, error) {
			return je.Void("actor-1", time.Now().Add(time.Second), "reason")
		},
	)

	require.Error(t, err)
}

func TestVoidJournalEntry_JournalEntryNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(100_000))
	nonExistentJEUUID := domain.JournalEntryUUID{UUID: common.NewUUIDv7()}

	_, err := repo.VoidJournalEntry(
		ctx,
		fs.UUID(),
		nonExistentJEUUID,
		func(_ *domain.FundSource, je *domain.JournalEntry) (*domain.JournalEntry, error) {
			return je.Void("actor-1", time.Now().Add(time.Second), "reason")
		},
	)

	require.Error(t, err)
}

func TestVoidJournalEntry_VoidFnError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutils.NewDB()
	repo := repoDb.NewFundSourceRepository(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))
	entryUUID := mustRecordDebitEntry(t, ctx, repo, fs.UUID(), decimal.NewFromInt(100_000))

	_, err := repo.VoidJournalEntry(
		ctx,
		fs.UUID(),
		entryUUID,
		func(_ *domain.FundSource, _ *domain.JournalEntry) (*domain.JournalEntry, error) {
			return nil, errors.New("voidFn error")
		},
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "voidFn error")

	original, err := dbmodels.New(db).GetJournalEntryByUID(ctx, entryUUID)
	require.NoError(t, err)
	assert.Equal(t, domain.JournalEntryStatusRecorded, original.Status)
}

func mustRecordDebitEntry(
	t *testing.T,
	ctx context.Context,
	repo *repoDb.FundSourceRepo,
	fsUUID domain.FundSourceUUID,
	amount decimal.Decimal,
) domain.JournalEntryUUID {
	t.Helper()

	var entryUUID domain.JournalEntryUUID

	err := repo.RecordJournalEntries(ctx, fsUUID, func(fs *domain.FundSource) ([]*domain.JournalEntry, error) {
		money, err := shared.NewMoney(amount, vnd)
		if err != nil {
			return nil, err
		}

		balanceBefore := fs.Balance().Amount()
		if err = fs.Withdraw(money); err != nil {
			return nil, err
		}

		entry, err := domain.NewJournalEntry(amount, shared.EntryTypeDebit, time.Now(), nil, nil, fsUUID, balanceBefore)
		if err != nil {
			return nil, err
		}
		entry.SetBalanceAfter(fs.Balance().Amount())

		entryUUID = entry.UUID()
		return []*domain.JournalEntry{entry}, nil
	})

	require.NoError(t, err)
	return entryUUID
}

func mustCreateCashFundSource(t *testing.T, ctx context.Context, repo *repoDb.FundSourceRepo, balanceAmt decimal.Decimal) *domain.FundSource {
	t.Helper()

	balance, err := shared.NewMoney(balanceAmt, vnd)
	require.NoError(t, err)

	metadata, err := fundSourceFactory.NewCashMetadata(gofakeit.Name())
	require.NoError(t, err)

	fs, err := fundSourceFactory.NewFundSource("Cash Wallet", domain.FundSourceTypeCash, balance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, repo.SaveFundSource(ctx, fs))
	return fs
}

func fundSourceToExpectedRow(fs *domain.FundSource) dbmodels.TreasuryFundSource {
	row := dbmodels.TreasuryFundSource{
		FundSourceUuid:   fs.UUID(),
		Name:             fs.Name(),
		SourceType:       fs.SourceType(),
		Balance:          fs.Balance().Amount(),
		AvailableBalance: fs.AvailableBalance().Amount(),
		Currency:         fs.Currency(),
		CreatedAt:        fs.CreatedAt(),
		CreatedBy:        fs.CreatedBy(),
		UpdatedAt:        fs.UpdatedAt(),
		UpdatedBy:        fs.UpdatedBy(),
	}

	if bankMeta, ok := fs.BankMetadata(); ok {
		row.BankInfo = bankMeta.BankInfo()
		row.BankCode = common.ToPtr(bankMeta.BankInfo().BankCode())
		row.BankAccountNumber = common.ToPtr(bankMeta.AccountNumber())
		row.BankAccountOwner = common.ToPtr(bankMeta.AccountOwner())
	}

	if cashMeta, ok := fs.CashMetadata(); ok {
		row.CashOwner = common.ToPtr(cashMeta.OwnerName())
	}

	return row
}
