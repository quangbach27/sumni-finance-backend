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

type ErrFundAllocatedNotFound struct {
	FpID string
}

func (e ErrFundAllocatedNotFound) Error() string {
	return fmt.Sprintf("fund provider: %s was not found", e.FpID)
}

type TransactionSpec struct {
	TransactionNo   string
	TransactionType string
	Amount          int64
	Description     string
	FpID            uuid.UUID
}

type Wallet struct {
	id      uuid.UUID
	name    string
	balance valueobject.Money
	version int32

	fpAllocationManager *FundProviderAllocationManager
	ledgerManager       *LedgerManager
}

// NewWallet constructs a new Wallet aggregate.
// Note: To rehydrate a Wallet from database state, use UnmarshalWalletFromDatabase, UnmarshalWalletWithLedgerFromDatabase instead
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

	fpAllocationManager, err := NewFpAllocationManager(nil)
	if err != nil {
		return nil, err
	}

	ledgerManager, err := NewLedgerManager(nil)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		id:                  id,
		name:                name,
		balance:             balance,
		version:             0,
		fpAllocationManager: fpAllocationManager,
		ledgerManager:       ledgerManager,
	}, nil
}

// UnmarshalWalletFromDatabase rehydrates a Wallet from persisted database state.
// The Wallet is initialized with provider allocations and an empty ledger manager.
// Note: Adding a new ledger manager afterward is supported and will not cause errors.
func UnmarshalWalletFromDatabase(
	id uuid.UUID,
	name string,
	balanceAmount int64,
	currencyCode string,
	version int32,
	providerAllocations ...*FpAllocation,
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

	providerManager, err := NewFpAllocationManager(providerAllocations)
	if err != nil {
		return nil, err
	}

	ledgerManager, err := NewLedgerManager(nil)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		id:                  id,
		name:                name,
		balance:             balance,
		version:             version,
		fpAllocationManager: providerManager,
		ledgerManager:       ledgerManager,
	}, nil
}

// UnmarshalWalletFromDatabase rehydrates a Wallet from persisted database state.
// The Wallet is initialized with provider allocations and ledger manager.
func UnmarshalWalletWithLedgerFromDatabase(
	id uuid.UUID,
	name string,
	balanceAmount int64,
	currencyCode string,
	version int32,
	accountingPeriods []*ledger.AccountingPeriod,
	providerAllocations ...*FpAllocation,
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

	ledgerManager, err := NewLedgerManager(accountingPeriods)
	if err != nil {
		return nil, err
	}

	w.ledgerManager = ledgerManager

	return w, nil
}

func (w *Wallet) ID() uuid.UUID                                       { return w.id }
func (w *Wallet) Name() string                                        { return w.name }
func (w *Wallet) Balance() valueobject.Money                          { return w.balance }
func (w *Wallet) Currency() valueobject.Currency                      { return w.balance.Currency() }
func (w *Wallet) Version() int32                                      { return w.version }
func (w *Wallet) FundProviderManager() *FundProviderAllocationManager { return w.fpAllocationManager }
func (w *Wallet) LedgerManager() *LedgerManager                       { return w.ledgerManager }

func (w *Wallet) SetAccountingPeriods(accountingPeriod ...*ledger.AccountingPeriod) error {
	ledgerManager, err := NewLedgerManager(accountingPeriod)
	if err != nil {
		return err
	}

	w.ledgerManager = ledgerManager
	return nil
}

// AllocateFundProvider registers a fund provider and increases the wallet balance by the allocated amount.
// The allocated amount represents the portion of the provider’s funds reserved for this wallet.
// It also updates the provider’s unallocated balance accordingly.
func (w *Wallet) AllocateFundProvider(
	fp *fundprovider.FundProvider,
	allocatedAmount int64,
) error {
	if allocatedAmount < 0 {
		return ErrAllocationAmountNegative
	}

	allocated, err := valueobject.NewMoney(allocatedAmount, w.Currency())
	if err != nil {
		return err
	}

	if err = w.fpAllocationManager.AddFundProviderAndReserve(fp, allocated); err != nil {
		return err
	}

	newWalletBalance, err := w.balance.Add(allocated)
	if err != nil {
		return err
	}

	w.balance = newWalletBalance
	return nil
}

func (w *Wallet) TopUp(amount valueobject.Money, fpID uuid.UUID) error {
	if amount.IsNegative() {
		return errors.New("amount must be positive")
	}

	allocation, exist := w.fpAllocationManager.FindFundProviderAllocation(fpID)
	if !exist {
		return ErrFundAllocatedNotFound{
			FpID: fpID.String(),
		}
	}

	if err := allocation.TopUpFundProviderAndAllocation(amount); err != nil {
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
	if amount.IsNegative() {
		return errors.New("amount must be positive")
	}

	allocation, exist := w.fpAllocationManager.FindFundProviderAllocation(fpID)
	if !exist {
		return ErrFundAllocatedNotFound{
			FpID: fpID.String(),
		}
	}

	if err := allocation.WithdrawFundProviderAndAllocation(amount); err != nil {
		return err
	}

	newWalletBalance, err := w.balance.Subtract(amount)
	if err != nil {
		return fmt.Errorf("withdraw in wallet failed: %w", err)
	}

	w.balance = newWalletBalance
	return nil
}

func (w *Wallet) OpenAccountingPeriod(yearMonth ledger.YearMonth) error {
	return w.ledgerManager.OpenNewAccountingPeriod(yearMonth, w.balance)
}

func (w *Wallet) RecordTransactions(yearMonth ledger.YearMonth, txSpecs ...TransactionSpec) error {
	if len(txSpecs) == 0 {
		return errors.New("transaction specs is empty")
	}

	accountingPeriod, exist := w.ledgerManager.FindAccountingPeriod(yearMonth)
	if !exist {
		return fmt.Errorf("account period: %s not found", yearMonth.String())
	}

	if accountingPeriod.IsClose() {
		return fmt.Errorf("accounting period: %s has been closed", yearMonth.String())
	}

	for _, txSpec := range txSpecs {
		txRecord, err := w.buildTransactionRecordsFromSpec(txSpec)
		if err != nil {
			return fmt.Errorf("failed to build transaction record: %w", err)
		}

		if err = w.ledgerManager.Record(yearMonth, txRecord); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) buildTransactionRecordsFromSpec(txSpec TransactionSpec) (ledger.TransactionRecord, error) {
	allocation, exist := w.fpAllocationManager.FindFundProviderAllocation(txSpec.FpID)
	if !exist {
		return ledger.TransactionRecord{}, ErrFundAllocatedNotFound{FpID: txSpec.FpID.String()}
	}
	amount, err := valueobject.NewMoney(txSpec.Amount, w.Currency())
	if err != nil {
		return ledger.TransactionRecord{}, err
	}

	if amount.Amount() <= 0 {
		return ledger.TransactionRecord{}, fmt.Errorf("transaction amount can not be negative or zero")
	}

	txRecord, err := ledger.NewTransactionRecord(
		txSpec.TransactionNo,
		txSpec.TransactionType,
		amount,
		txSpec.Description,
		txSpec.FpID,
	)
	if err != nil {
		return ledger.TransactionRecord{}, err
	}

	if txRecord.IsCredit() {
		if err = w.TopUp(txRecord.Amount(), txSpec.FpID); err != nil {
			return ledger.TransactionRecord{}, err
		}
	} else {
		if err = w.Withdraw(txRecord.Amount(), txSpec.FpID); err != nil {
			return ledger.TransactionRecord{}, err
		}
	}

	txRecord.SetFpBalance(allocation.FundProvider().Balance())
	txRecord.SetWalletBalance(w.balance)

	return *txRecord, nil
}
