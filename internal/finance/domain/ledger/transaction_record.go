package ledger

import (
	"fmt"
	"strings"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"

	"github.com/google/uuid"
)

var (
	DebitTransaction  TransactionType = TransactionType{"DEBIT"}
	CreditTransaction TransactionType = TransactionType{"CREDIT"}
)

type TransactionType struct {
	value string
}

func (t TransactionType) String() string { return t.value }

func NewTransactionType(typeStr string) (TransactionType, error) {
	typeStrCleaned := strings.ToUpper(strings.TrimSpace(typeStr))
	if typeStrCleaned == DebitTransaction.value {
		return DebitTransaction, nil
	}

	if typeStrCleaned == CreditTransaction.value {
		return CreditTransaction, nil
	}

	return TransactionType{}, fmt.Errorf("unknown transaction type: %s", typeStr)
}

type TransactionRecord struct {
	id              uuid.UUID
	transactionNo   string
	transactionType TransactionType // debit, credit
	amount          valueobject.Money
	description     string

	walletBalance valueobject.Money
	fpID          uuid.UUID
	fpBalance     valueobject.Money
}

func NewTransactionRecord(
	transactionNo string,
	transactionType string,
	amount valueobject.Money,
	description string,
	fpID uuid.UUID,
) (*TransactionRecord, error) {
	v := validator.New()

	v.Required(transactionType, "transactionType")
	v.Check(!amount.IsZero(), "amount", "amount is required")
	v.Check(amount.Amount() > 0, "amount", "amount must be positive")
	v.Check(fpID != uuid.Nil, "fpID", "fpID is required")

	if err := v.Err(); err != nil {
		return nil, fmt.Errorf("new transaction record: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("new transaction record: %w", err)
	}

	txType, err := NewTransactionType(transactionType)
	if err != nil {
		return nil, fmt.Errorf("new transaction record: %w", err)
	}

	return &TransactionRecord{
		id:              id,
		amount:          amount,
		transactionNo:   transactionNo,
		transactionType: txType,
		description:     description,
		fpID:            fpID,
	}, nil
}

func (tr *TransactionRecord) SetWalletBalance(walletBalance valueobject.Money) {
	tr.walletBalance = walletBalance
}

func (tr *TransactionRecord) SetFpBalance(fpBalance valueobject.Money) {
	tr.fpBalance = fpBalance
}

func (t *TransactionRecord) ID() uuid.UUID                    { return t.id }
func (t *TransactionRecord) TransactionNo() string            { return t.transactionNo }
func (t *TransactionRecord) TransactionType() TransactionType { return t.transactionType }
func (t *TransactionRecord) Amount() valueobject.Money        { return t.amount }
func (t *TransactionRecord) Description() string              { return t.description }
func (t *TransactionRecord) WalletBalance() valueobject.Money { return t.walletBalance }
func (t *TransactionRecord) FpID() uuid.UUID                  { return t.fpID }
func (t *TransactionRecord) FpBalance() valueobject.Money     { return t.fpBalance }

func (t *TransactionRecord) IsCredit() bool {
	return t.transactionType.value == CreditTransaction.value
}
