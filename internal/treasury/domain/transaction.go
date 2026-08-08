package domain

import (
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

// TransactionType represents the lifecycle state of a transaction. A transaction
// moves through these states in order, generally as follows:
//
//	DRAFTED -> RECORDED -> POSTED
//
// There are 5 states:
//
//  1. DRAFTED
//     - The transaction is stored in the DB with a fund source UUID only.
//     - The fund source balance is NOT yet affected.
//     - The transaction is editable in this state.
//
//  2. RECORDED
//     - The transaction is stored in the DB with both a fund source UUID
//     and a wallet UUID.
//     - The wallet and fund source balances ARE affected by the transaction
//     amount.
//     - The transaction is still editable in this state; editing it updates
//     the wallet and fund source balances accordingly.
//
//  3. POSTED
//     - The transaction has been posted to the ledger and is now immutable.
//     - To correct a POSTED transaction, the user must void it (see VOIDED)
//     and create a new transaction rather than editing it directly.
//
//  4. VOIDED
//     - The POSTED transaction has been marked as void. This records the
//     intent to cancel its effect but does not by itself move any money.
//     - A VOIDED transaction must be paired with a corresponding REVERSED
//     transaction (see below) before its financial effect is actually
//     cancelled out.
//
//  5. REVERSED
//     - A new transaction has been created to offset a VOIDED transaction,
//     such that the sum of the VOIDED transaction and its REVERSED
//     counterpart is zero.
//     - This is the state that actually neutralizes the original
//     transaction's effect on wallet/fund source balances.
type TransactionType struct {
	common.Enum[TransactionTypeValues]
}

type TransactionTypeValues string

func (tv TransactionTypeValues) Values() []string {
	return []string{"DRAFTED", "RECORDED", "POSTED", "VOIDED", "REVERSED"}
}

var (
	TransactionTypeDrafted  = common.MustEnum[TransactionType]("DRAFTED")
	TransactionTypeRecorded = common.MustEnum[TransactionType]("RECORDED")
	TransactionTypePosted   = common.MustEnum[TransactionType]("POSTED")
	TransactionTypeVoided   = common.MustEnum[TransactionType]("VOIDED")
	TransactionTypeReversed = common.MustEnum[TransactionType]("REVERSED")
)

type TransactionUUID struct {
	common.UUID
}
type Transaction struct {
	uuid TransactionUUID

	transactionType TransactionType
	entryType       shared.EntryType
	amount          shared.Money

	fundSourceUUID          FundSourceUUID
	walletUUID              *WalletUUID
	reversedTransactionUUID *TransactionUUID
}

func (t *Transaction) UUID() TransactionUUID                     { return t.uuid }
func (t *Transaction) TransactionType() TransactionType          { return t.transactionType }
func (t *Transaction) EntryType() shared.EntryType               { return t.entryType }
func (t *Transaction) Amount() shared.Money                      { return t.amount }
func (t *Transaction) FundSourceUUID() FundSourceUUID            { return t.fundSourceUUID }
func (t *Transaction) WalletUUID() *WalletUUID                   { return t.walletUUID }
func (t *Transaction) ReversedTransactionUUID() *TransactionUUID { return t.reversedTransactionUUID }

func NewTransaction(
	entryType shared.EntryType,
	amount shared.Money,
	fundSourceUUID FundSourceUUID,
	walletUUID WalletUUID,
) (*Transaction, error) {
	errDetails := []common.ErrorDetails{}
	if entryType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Transaction",
			ErrorSlug:  "empty-entry-type",
			Message:    "entry type can't be empty",
		})
	}
	if !amount.Amount().IsPositive() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Transaction",
			ErrorSlug:  "invalid-amount",
			Message:    "transaction amount must be positve",
		})
	}
	if fundSourceUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Transaction",
			ErrorSlug:  "empty-fund-source-uuid",
			Message:    "fund source uuid can't be empty",
		})
	}
	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-transaction", "transacion is not valid").WithDetails(errDetails)
	}

	transactionType := TransactionTypeDrafted
	if !walletUUID.IsZero() {
		transactionType = TransactionTypeRecorded
	}

	var walletUUIDPtr *WalletUUID
	if !walletUUID.IsZero() {
		walletUUIDPtr = &walletUUID
	}

	return &Transaction{
		uuid:            TransactionUUID{common.NewUUIDv7()},
		transactionType: transactionType,
		entryType:       entryType,
		amount:          amount,
		fundSourceUUID:  fundSourceUUID,
		walletUUID:      walletUUIDPtr,
	}, nil
}

func (t *Transaction) MarkAsRecorded() error {
	if t.transactionType != TransactionTypeDrafted {
		return fmt.Errorf("can't mark as record with transaction type: %s", t.transactionType.String())
	}

	t.transactionType = TransactionTypeRecorded
	return nil
}

func (t *Transaction) MarkAsPosted() error {
	if t.transactionType != TransactionTypeRecorded {
		return fmt.Errorf("can't mark as post with transaction type: %s", t.transactionType.String())
	}

	t.transactionType = TransactionTypePosted
	return nil
}

func (t *Transaction) Void() (*Transaction, error) {
	if t.transactionType != TransactionTypePosted {
		return nil, fmt.Errorf("")
	}
	reverseTransaction := t.newReverseTransaction()

	t.transactionType = TransactionTypeVoided
	t.reversedTransactionUUID = common.ToPtr(reverseTransaction.uuid)

	return reverseTransaction, nil
}

func (t *Transaction) newReverseTransaction() *Transaction {
	var reverseEntryType shared.EntryType
	if t.entryType == shared.EntryTypeCredit {
		reverseEntryType = shared.EntryTypeDebit
	} else {
		reverseEntryType = shared.EntryTypeCredit
	}

	return &Transaction{
		uuid:            TransactionUUID{common.NewUUIDv7()},
		transactionType: TransactionTypeReversed,
		entryType:       reverseEntryType,
		amount:          t.amount,
		fundSourceUUID:  t.fundSourceUUID,
		walletUUID:      t.walletUUID,
	}
}
