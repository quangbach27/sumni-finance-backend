package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
)

func TestNewTransaction(t *testing.T) {
	t.Parallel()

	validAmount, err := shared.NewMoney(decimal.NewFromInt(100_000), shared.MustNewCurrency("VND"))
	require.NoError(t, err)

	validFundSourceUUID := domain.FundSourceUUID{UUID: common.NewUUIDv7()}
	validWalletUUID := domain.WalletUUID{UUID: common.NewUUIDv7()}

	tests := []struct {
		name            string
		entryType       shared.EntryType
		amount          shared.Money
		fundSourceUUID  domain.FundSourceUUID
		walletUUID      domain.WalletUUID
		wantType        domain.TransactionType
		wantErrContains string
	}{
		{
			name:           "creates recorded transaction when wallet exists",
			entryType:      shared.EntryTypeDebit,
			amount:         validAmount,
			fundSourceUUID: validFundSourceUUID,
			walletUUID:     validWalletUUID,
			wantType:       domain.TransactionTypeRecorded,
		},
		{
			name:           "creates drafted transaction when wallet is empty",
			entryType:      shared.EntryTypeCredit,
			amount:         validAmount,
			fundSourceUUID: validFundSourceUUID,
			wantType:       domain.TransactionTypeDrafted,
		},
		{
			name:            "returns error when entry type is empty",
			entryType:       shared.EntryType{},
			amount:          validAmount,
			fundSourceUUID:  validFundSourceUUID,
			wantErrContains: "empty-entry-type",
		},
		{
			name:            "returns error when amount is not positive",
			entryType:       shared.EntryTypeDebit,
			amount:          newValidMoney(t, 0, vnd),
			fundSourceUUID:  validFundSourceUUID,
			wantErrContains: "invalid-amount",
		},
		{
			name:            "returns error when fund source uuid is empty",
			entryType:       shared.EntryTypeDebit,
			amount:          validAmount,
			fundSourceUUID:  domain.FundSourceUUID{},
			wantErrContains: "empty-fund-source-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transaction, err := domain.NewTransaction(tt.entryType, tt.amount, tt.fundSourceUUID, tt.walletUUID)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, transaction)
			assert.Equal(t, tt.wantType, transaction.TransactionType())
			assert.Equal(t, tt.entryType, transaction.EntryType())
			assert.Equal(t, tt.amount, transaction.Amount())
			if tt.wantType == domain.TransactionTypeDrafted {
				assert.Nil(t, transaction.WalletUUID())
			} else {
				require.NotNil(t, transaction.WalletUUID())
				assert.Equal(t, tt.walletUUID, *transaction.WalletUUID())
			}
			assert.Equal(t, tt.fundSourceUUID, transaction.FundSourceUUID())
		})
	}
}

func TestTransaction_StateTransitions(t *testing.T) {
	t.Parallel()

	amount, err := shared.NewMoney(decimal.NewFromInt(250_000), shared.MustNewCurrency("VND"))
	require.NoError(t, err)

	fundSourceUUID := domain.FundSourceUUID{UUID: common.NewUUIDv7()}
	walletUUID := domain.WalletUUID{UUID: common.NewUUIDv7()}

	t.Run("mark as recorded changes drafted transaction to recorded", func(t *testing.T) {
		t.Parallel()

		transaction, err := domain.NewTransaction(shared.EntryTypeDebit, amount, fundSourceUUID, domain.WalletUUID{})
		require.NoError(t, err)

		err = transaction.MarkAsRecorded()
		require.NoError(t, err)
		assert.Equal(t, domain.TransactionTypeRecorded, transaction.TransactionType())
	})

	t.Run("mark as recorded returns error for non-drafted transaction", func(t *testing.T) {
		t.Parallel()

		transaction, err := domain.NewTransaction(shared.EntryTypeCredit, amount, fundSourceUUID, walletUUID)
		require.NoError(t, err)

		err = transaction.MarkAsRecorded()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't mark as record")
	})

	t.Run("mark as posted returns error for non-recorded transaction", func(t *testing.T) {
		t.Parallel()

		transaction, err := domain.NewTransaction(shared.EntryTypeDebit, amount, fundSourceUUID, domain.WalletUUID{})
		require.NoError(t, err)

		err = transaction.MarkAsPosted()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't mark as post")
	})

	t.Run("void returns a reversed transaction for posted transaction", func(t *testing.T) {
		t.Parallel()

		transaction, err := domain.NewTransaction(shared.EntryTypeDebit, amount, fundSourceUUID, walletUUID)
		require.NoError(t, err)
		require.NoError(t, transaction.MarkAsPosted())

		reverseTransaction, err := transaction.Void()
		require.NoError(t, err)
		require.NotNil(t, reverseTransaction)
		assert.Equal(t, domain.TransactionTypeVoided, transaction.TransactionType())
		assert.Equal(t, domain.TransactionTypeReversed, reverseTransaction.TransactionType())
		assert.Equal(t, shared.EntryTypeCredit, reverseTransaction.EntryType())
		assert.Equal(t, amount, reverseTransaction.Amount())
		assert.Equal(t, fundSourceUUID, reverseTransaction.FundSourceUUID())
		assert.Equal(t, transaction.WalletUUID(), reverseTransaction.WalletUUID())
		assert.Equal(t, reverseTransaction.UUID(), *transaction.ReversedTransactionUUID())
	})
}
