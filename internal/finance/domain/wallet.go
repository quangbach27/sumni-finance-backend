package domain

import (
	"errors"
	"fmt"
	"strings"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

type AllocatableFundProvider interface {
	UUID() FundProviderUUID
	Balance() shared.Money
	reserve(m shared.Money) error
	topUp(m shared.Money) error
	withdraw(m shared.Money) error
}

type WalletUUID struct {
	common.UUID
}

type Wallet struct {
	uuid        WalletUUID
	name        string
	description string

	balance  shared.Money
	currency shared.Currency

	fundProviderRegister *fundProviderRegistry
}

func (w *Wallet) UUID() WalletUUID {
	return w.uuid
}

func (w *Wallet) Name() string {
	return w.name
}

func (w *Wallet) Description() string {
	return w.description
}

func (w *Wallet) Balance() shared.Money {
	return w.balance
}

func (w *Wallet) Currency() shared.Currency {
	return w.currency
}

type WalletAllocationSnapshot struct {
	WalletUUID    WalletUUID
	WalletBalance shared.Money

	WalletAllocation shared.Money

	FundProviderUUID    FundProviderUUID
	FundProviderBalance shared.Money
}

func (w *Wallet) TopUp(amount shared.Money, fpUUID FundProviderUUID) (WalletAllocationSnapshot, error) {
	if amount.IsZero() {
		return WalletAllocationSnapshot{}, errors.New("money for top up can't be empty")
	}

	if amount.Amount().IsZero() || amount.Amount().IsNegative() {
		return WalletAllocationSnapshot{}, errors.New("amount for top up must be positive")
	}

	if !w.currency.Equal(amount.Currency()) {
		return WalletAllocationSnapshot{}, fmt.Errorf("top up amount currency %s does not match with wallet currency %s", amount.Currency().String(), w.currency.String())
	}

	fpBalance, allocationAmount, err := w.fundProviderRegister.increaseAllocation(fpUUID, amount)
	if err != nil {
		return WalletAllocationSnapshot{}, err
	}

	newBalance, err := w.balance.Add(amount)
	if err != nil {
		return WalletAllocationSnapshot{}, err
	}

	w.balance = newBalance

	return WalletAllocationSnapshot{
		WalletUUID:          w.uuid,
		WalletBalance:       w.balance,
		WalletAllocation:    allocationAmount,
		FundProviderUUID:    fpUUID,
		FundProviderBalance: fpBalance,
	}, nil
}

func (w *Wallet) Withdraw(amount shared.Money, fpUUID FundProviderUUID) (WalletAllocationSnapshot, error) {
	if amount.IsZero() {
		return WalletAllocationSnapshot{}, errors.New("money for top up can't be empty")
	}

	if amount.Amount().IsZero() || amount.Amount().IsNegative() {
		return WalletAllocationSnapshot{}, errors.New("amount for top up must be positive")
	}

	if !w.currency.Equal(amount.Currency()) {
		return WalletAllocationSnapshot{}, fmt.Errorf("top up amount currency %s does not match with wallet currency %s", amount.Currency().String(), w.currency.String())
	}

	if amount.Amount().GreaterThan(w.balance.Amount()) {
		return WalletAllocationSnapshot{}, fmt.Errorf("withdraw amount must be less then wallet balance")
	}

	fpBalance, allocationAmount, err := w.fundProviderRegister.decreaseAllocation(fpUUID, amount)
	if err != nil {
		return WalletAllocationSnapshot{}, err
	}

	newBalance, err := w.balance.Sub(amount)
	if err != nil {
		return WalletAllocationSnapshot{}, err
	}
	w.balance = newBalance

	return WalletAllocationSnapshot{
		WalletUUID:          w.uuid,
		WalletBalance:       w.balance,
		WalletAllocation:    allocationAmount,
		FundProviderUUID:    fpUUID,
		FundProviderBalance: fpBalance,
	}, nil
}

func (w *Wallet) AllocateFundProvider(fp AllocatableFundProvider, amount shared.Money) error {
	err := w.fundProviderRegister.register(fp, amount)
	if err != nil {
		return err
	}

	newBalance, err := w.balance.Add(amount)
	if err != nil {
		return err
	}

	w.balance = newBalance

	return nil
}

type NewFundProviderAllocationData struct {
	FundProvider AllocatableFundProvider
	Amount       shared.Money
}

func NewWallet(
	name string,
	description string,
	currency shared.Currency,
	allocationsData []NewFundProviderAllocationData,
) (*Wallet, error) {
	errDetails := []common.ErrorDetails{}
	if strings.TrimSpace(name) == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "wallet",
			ErrorSlug:  "empty-name",
			Message:    "name can't not be empty",
		})
	}

	if currency.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "wallet",
			ErrorSlug:  "empty-currency",
			Message:    "currency can't not be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-wallet-input", "invalid wallet input").WithDetails(errDetails)
	}

	allocationManager, err := newFundProviderRegistry(currency, allocationsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocations manager: %w", err)
	}

	balance, err := shared.NewMoney(decimal.Zero, currency)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocationsData {
		balance, err = balance.Add(allocation.Amount)
		if err != nil {
			return nil, err
		}
	}

	return &Wallet{
		uuid:                 WalletUUID{UUID: common.NewUUIDv7()},
		name:                 name,
		description:          description,
		balance:              balance,
		currency:             currency,
		fundProviderRegister: allocationManager,
	}, nil
}

type fundProviderRegistry struct {
	currency    shared.Currency
	allocations map[FundProviderUUID]*fundProviderAllocation
}

type fundProviderAllocation struct {
	fundProvider AllocatableFundProvider
	amount       shared.Money
}

func newFundProviderRegistry(currency shared.Currency, allocationsData []NewFundProviderAllocationData) (*fundProviderRegistry, error) {
	if currency.IsZero() {
		return nil, errors.New("currency can't be empty")
	}

	allocations := make(map[FundProviderUUID]*fundProviderAllocation, len(allocationsData))

	for _, data := range allocationsData {
		fp := data.FundProvider
		if fp == nil {
			return nil, errors.New("fund provider can't be empty")
		}

		if _, ok := allocations[fp.UUID()]; ok {
			return nil, fmt.Errorf("duplicate fund provider '%s' is not allowed", fp.UUID())
		}

		err := data.FundProvider.reserve(data.Amount)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to allocate fund provider %s with amount %s: %w",
				data.FundProvider.UUID().String(),
				data.Amount.String(),
				err,
			)
		}

		allocations[data.FundProvider.UUID()] = &fundProviderAllocation{
			fundProvider: data.FundProvider,
			amount:       data.Amount,
		}
	}

	return &fundProviderRegistry{
		currency:    currency,
		allocations: allocations,
	}, nil
}

func (m *fundProviderRegistry) hasFundProvider(fpUUID FundProviderUUID) bool {
	_, ok := m.allocations[fpUUID]
	return ok
}

func (m *fundProviderRegistry) register(fp AllocatableFundProvider, amount shared.Money) error {
	if fp == nil {
		return errors.New("fund provider can't be empty")
	}

	if amount.IsZero() {
		return errors.New("amount can't be empty")
	}

	if m.hasFundProvider(fp.UUID()) {
		return fmt.Errorf("fund provider %s has already registered", fp.UUID().String())
	}

	if !m.currency.Equal(amount.Currency()) {
		return fmt.Errorf("amount currency %s does not match with wallet currency %s", amount.Currency().Code(), m.currency.Code())
	}

	err := fp.reserve(amount)
	if err != nil {
		return fmt.Errorf("failed to reserve fund provider: %w", err)
	}

	m.allocations[fp.UUID()] = &fundProviderAllocation{
		fundProvider: fp,
		amount:       amount,
	}

	return nil
}

func (m *fundProviderRegistry) increaseAllocation(fpUUID FundProviderUUID, amount shared.Money) (shared.Money, shared.Money, error) {
	if !m.hasFundProvider(fpUUID) {
		return shared.Money{}, shared.Money{}, errors.New("fund provider is not registered")
	}

	allocation := m.allocations[fpUUID]

	err := allocation.fundProvider.topUp(amount)
	if err != nil {
		return shared.Money{}, shared.Money{}, err
	}

	newAllocationAmount, err := allocation.amount.Add(amount)
	if err != nil {
		return shared.Money{}, shared.Money{}, err
	}

	allocation.amount = newAllocationAmount

	return allocation.fundProvider.Balance(), allocation.amount, nil
}

func (m *fundProviderRegistry) decreaseAllocation(fpUUID FundProviderUUID, amount shared.Money) (shared.Money, shared.Money, error) {
	if !m.hasFundProvider(fpUUID) {
		return shared.Money{}, shared.Money{}, errors.New("fund provider  is not registered")
	}

	allocation := m.allocations[fpUUID]

	if amount.Amount().GreaterThan(allocation.amount.Amount()) {
		return shared.Money{}, shared.Money{}, errors.New("allocation amount can't exceed wallet allocation")
	}

	newAllocationAmount, err := allocation.amount.Sub(amount)
	if err != nil {
		return shared.Money{}, shared.Money{}, err
	}
	allocation.amount = newAllocationAmount

	err = allocation.fundProvider.withdraw(amount)
	if err != nil {
		return shared.Money{}, shared.Money{}, fmt.Errorf("failed withdraw fund provider: %w", err)
	}

	return allocation.fundProvider.Balance(), allocation.amount, nil
}

type AccountingPeriodConfig struct {
	intervalInMonths int
	dayOfMonth       int
}

func (c AccountingPeriodConfig) IntervalInMonths() int {
	return c.intervalInMonths
}

func (c AccountingPeriodConfig) DayOfMonth() int {
	return c.dayOfMonth
}

func (c AccountingPeriodConfig) IsZero() bool {
	return c == AccountingPeriodConfig{}
}

func NewAccountingPeriodConfig(
	intervalInMonths,
	dayOfMonth int,
) (AccountingPeriodConfig, error) {
	if dayOfMonth < 1 || dayOfMonth > 27 {
		return AccountingPeriodConfig{}, errors.New("day of month must be between 1 and 27")
	}

	if intervalInMonths < 1 || intervalInMonths > 5 {
		return AccountingPeriodConfig{}, errors.New("interval in months must be between 1 and 5")
	}

	return AccountingPeriodConfig{
		intervalInMonths: intervalInMonths,
		dayOfMonth:       dayOfMonth,
	}, nil
}
