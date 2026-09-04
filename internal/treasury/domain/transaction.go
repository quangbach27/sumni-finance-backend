// Package domain
package domain

import (
	"errors"
	"fmt"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

// TransactionStatus represents the lifecycle state of a transaction. A transaction
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
type TransactionStatus struct {
	common.Enum[TransactionStatusValues]
}

type TransactionStatusValues string

func (tv TransactionStatusValues) Values() []string {
	return []string{"DRAFTED", "RECORDED", "POSTED", "VOIDED", "REVERSED"}
}

var (
	TransactionStatusDrafted  = common.MustEnum[TransactionStatus]("DRAFTED")
	TransactionStatusRecorded = common.MustEnum[TransactionStatus]("RECORDED")
	TransactionStatusPosted   = common.MustEnum[TransactionStatus]("POSTED")
	TransactionStatusVoided   = common.MustEnum[TransactionStatus]("VOIDED")
	TransactionStatusReversed = common.MustEnum[TransactionStatus]("REVERSED")
)

type TransactionUUID struct {
	common.UUID
}

type Transaction struct {
	uuid TransactionUUID

	status    TransactionStatus
	entryType shared.EntryType
	amount    shared.Money

	fundSourceUUID          FundSourceUUID
	walletUUID              *WalletUUID
	reversedTransactionUUID *TransactionUUID

	description     *string
	transactionDate time.Time
}

func (t *Transaction) UUID() TransactionUUID                     { return t.uuid }
func (t *Transaction) Status() TransactionStatus                 { return t.status }
func (t *Transaction) EntryType() shared.EntryType               { return t.entryType }
func (t *Transaction) Amount() shared.Money                      { return t.amount }
func (t *Transaction) FundSourceUUID() FundSourceUUID            { return t.fundSourceUUID }
func (t *Transaction) WalletUUID() *WalletUUID                   { return t.walletUUID }
func (t *Transaction) ReversedTransactionUUID() *TransactionUUID { return t.reversedTransactionUUID }
func (t *Transaction) Description() *string                      { return t.description }
func (t *Transaction) TransactionDate() time.Time                { return t.transactionDate }

func NewTransaction(
	entryType shared.EntryType,
	amount shared.Money,
	fundSourceUUID FundSourceUUID,
	walletUUID WalletUUID,
	description *string,
	transactionDate time.Time,
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
	if transactionDate.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Transaction",
			ErrorSlug:  "empty-transaction-date",
			Message:    "transaction date can't be empty",
		})
	}
	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-transaction", "transacion is not valid").WithDetails(errDetails)
	}

	var walletUUIDPtr *WalletUUID
	status := TransactionStatusDrafted

	if !walletUUID.IsZero() {
		walletUUIDPtr = &walletUUID
		status = TransactionStatusRecorded
	}

	return &Transaction{
		uuid:            TransactionUUID{common.NewUUIDv7()},
		status:          status,
		entryType:       entryType,
		amount:          amount,
		fundSourceUUID:  fundSourceUUID,
		walletUUID:      walletUUIDPtr,
		description:     description,
		transactionDate: transactionDate,
	}, nil
}

func (t *Transaction) IsRecorded() bool {
	return t.status == TransactionStatusRecorded
}

func (t *Transaction) MarkAsRecorded(walletUUID WalletUUID) error {
	if walletUUID.IsZero() {
		return errors.New("walletUUID can't be empty")
	}

	if t.status != TransactionStatusDrafted {
		return fmt.Errorf("can't mark as record with transaction type: %s", t.status.String())
	}

	t.status = TransactionStatusRecorded
	t.walletUUID = &walletUUID

	return nil
}

func (t *Transaction) MarkAsPosted() error {
	if t.status != TransactionStatusRecorded {
		return fmt.Errorf("can't mark as post with transaction type: %s", t.status.String())
	}

	t.status = TransactionStatusPosted
	return nil
}

func (t *Transaction) Void() (*Transaction, error) {
	if t.status != TransactionStatusPosted {
		return nil, fmt.Errorf("transaction can't be voided because transaction (status: %s) isn't posted", t.status.String())
	}
	reverseTransaction := t.newReverseTransaction()

	t.status = TransactionStatusVoided
	t.reversedTransactionUUID = common.ToPtr(reverseTransaction.uuid)

	return reverseTransaction, nil
}

func (t *Transaction) newReverseTransaction() *Transaction {
	var reverseEntryType shared.EntryType
	if t.entryType == shared.EntryTypeOut {
		reverseEntryType = shared.EntryTypeIn
	} else {
		reverseEntryType = shared.EntryTypeOut
	}

	return &Transaction{
		uuid:            TransactionUUID{common.NewUUIDv7()},
		status:          TransactionStatusReversed,
		entryType:       reverseEntryType,
		amount:          t.amount,
		fundSourceUUID:  t.fundSourceUUID,
		walletUUID:      t.walletUUID,
		description:     t.description,
		transactionDate: time.Now(),
	}
}
