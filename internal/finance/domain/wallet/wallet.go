package wallet

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"sumni-finance-backend/internal/finance/domain/ledger"

	"github.com/google/uuid"
)

var (
	ErrFundProviderAlreadyRegistered = errors.New("fund provider already registered")
	ErrAllocationAmountNegative      = errors.New("allocated amount is negative")
)

type ErrFundAllocatedMissing struct {
	FpID string
}

func (e ErrFundAllocatedMissing) Error() string {
	return fmt.Sprintf("fund provider: %s is missing", e.FpID)
}

type TransactionSpec struct {
	year  int
	month int

	TransactionNo   string
	TransactionType string
	Amount          valueobject.Money
	Description     string
	fpID            uuid.UUID
}

type Wallet struct {
	id      uuid.UUID
	name    string
	balance valueobject.Money
	version int32

	providerManager *ProviderManager
	ledgerManager   *LedgerManager
}

func NewWallet(currencyCode string, name string) (*Wallet, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	currency, err := valueobject.NewCurrency(currencyCode)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	balance, err := valueobject.NewMoney(0, currency)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		id:      id,
		name:    name,
		balance: balance,
		version: 0,
		providerManager: &ProviderManager{
			providers: make(map[uuid.UUID]ProviderAllocation),
		},
	}, nil
}

func UnmarshalWalletFromDatabase(
	id uuid.UUID,
	name string,
	balanceAmount int64,
	currencyCode string,
	version int32,
	providerAllocations ...ProviderAllocation,
) (*Wallet, error) {
	v := validator.New()

	v.Check(id != uuid.Nil, "id", "id is required")
	v.Required(name, "name")
	v.Check(balanceAmount >= 0, "balance", "balance must greater or equal than 0")
	v.Required(currencyCode, "currency")

	if err := v.Err(); err != nil {
		return nil, err
	}

	currency, err := valueobject.NewCurrency(currencyCode)
	if err != nil {
		return nil, err
	}

	balance, err := valueobject.NewMoney(balanceAmount, currency)
	if err != nil {
		return nil, err
	}

	providerManager, err := NewProviderManager(providerAllocations)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		id:              id,
		name:            name,
		balance:         balance,
		version:         version,
		providerManager: providerManager,
	}, nil
}

func UmarshalWalletFromDatabaseWithLedger(
	id uuid.UUID,
	name string,
	balanceAmount int64,
	currencyCode string,
	version int32,
	accountingPerids []*ledger.AccountingPeriod,
	providerAllocations ...ProviderAllocation,
) (*Wallet, error) {
	w, err := UnmarshalWalletFromDatabase(
		id,
		name,
		balanceAmount,
		currencyCode,
		version,
		providerAllocations...,
	)
	if err != nil {
		return nil, err
	}

	ledgerManger, err := NewLedgerManager(accountingPerids)
	if err != nil {
		return nil, err
	}

	w.ledgerManager = ledgerManger

	return w, nil
}

func (w *Wallet) ID() uuid.UUID                     { return w.id }
func (w *Wallet) Name() string                      { return w.name }
func (w *Wallet) Balance() valueobject.Money        { return w.balance }
func (w *Wallet) Currency() valueobject.Currency    { return w.balance.Currency() }
func (w *Wallet) Version() int32                    { return w.version }
func (w *Wallet) ProviderManager() *ProviderManager { return w.providerManager }

func (w *Wallet) AllocateFromFundProvider(
	fundProvider *fundprovider.FundProvider,
	allocatedAmount int64,
) error {
	if fundProvider == nil {
		return ErrFundAllocatedMissing{
			FpID: "unknown",
		}
	}

	if allocatedAmount < 0 {
		return ErrAllocationAmountNegative
	}

	if fp := w.ProviderManager().FindProvider(fundProvider.ID()); fp != nil {
		return ErrFundProviderAlreadyRegistered
	}

	allocated, err := valueobject.NewMoney(allocatedAmount, w.Currency())
	if err != nil {
		return err
	}

	if err = w.providerManager.AddFundProviderAndReserve(fundProvider, allocated); err != nil {
		return err
	}

	newWalletBalance, err := w.balance.Add(allocated)
	if err != nil {
		return err
	}

	w.balance = newWalletBalance
	return nil
}

func (w *Wallet) Topup(amount valueobject.Money, fpID uuid.UUID) error {
	fp := w.providerManager.FindProvider(fpID)
	if fp == nil {
		return ErrFundAllocatedMissing{
			FpID: fpID.String(),
		}
	}

	if err := fp.TopUp(amount); err != nil {
		return err
	}

	newWalletBalance, err := w.balance.Add(amount)
	if err != nil {
		return fmt.Errorf("topup in wallet failed: %w", err)
	}

	w.balance = newWalletBalance
	return nil
}

func (w *Wallet) Withdraw(amount valueobject.Money, fpID uuid.UUID) error {
	fp := w.providerManager.FindProvider(fpID)
	if fp == nil {
		return ErrFundAllocatedMissing{
			FpID: fpID.String(),
		}
	}

	if err := fp.Withdraw(amount); err != nil {
		return err
	}

	newWalletBalance, err := w.balance.Subtract(amount)
	if err != nil {
		return fmt.Errorf("withdraw in wallet failed: %w", err)
	}

	w.balance = newWalletBalance
	return nil
}

func (w *Wallet) OpenAccountPeriod(yearMonth ledger.YearMonth) error {
	return w.ledgerManager.OpenNewAccountingPeriod(yearMonth, w.balance)
}

func (w *Wallet) RecordTransactions(yearMonth ledger.YearMonth, specs ...TransactionSpec) error {
	accountingPeriod := w.ledgerManager.accountPeriods[yearMonth]

	if accountingPeriod == nil || accountingPeriod.IsClose() {
		return fmt.Errorf("account period: %s has been closed", yearMonth.String())
	}

	transactionRecords := make([]*ledger.TransactionRecord, 0, len(specs))

	for _, spec := range specs {
		transactionRecord, err := ledger.NewTransactionRecord(
			spec.TransactionNo,
			spec.TransactionType,
			spec.Amount,
			spec.Description,
			spec.fpID,
		)
		if err != nil {
			return fmt.Errorf("record transaction: %w", err)
		}

		if transactionRecord.IsCredit() {
			err = w.Topup(transactionRecord.Amount(), spec.fpID)
			if err != nil {
				return fmt.Errorf("record transaction failed: %w", err)
			}
		} else {
			err = w.Withdraw(transactionRecord.Amount(), spec.fpID)
			if err != nil {
				return fmt.Errorf("record transaction failed: %w", err)
			}
		}

		transactionRecord.SetFpBalance(w.providerManager.FindProvider(spec.fpID).Balance())
		transactionRecord.SetWalletBalance(w.balance)

		transactionRecords = append(transactionRecords, transactionRecord)
	}

	if err := w.ledgerManager.Record(yearMonth, transactionRecords); err != nil {
		return fmt.Errorf("record transaction failed: %w", err)
	}

	return nil
}
