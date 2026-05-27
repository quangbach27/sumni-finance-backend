package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/app/models"

	"github.com/shopspring/decimal"
)

type WalletRepository interface {
	GetWalletWithAllocations(ctx context.Context, walletUUID WalletUUID) (*Wallet, error)

	SaveWallet(
		ctx context.Context,
		fundProvidersUUIDs []FundProviderUUID,
		createFn func(
			fundProviders map[FundProviderUUID]*FundProvider,
		) (*Wallet, *Ledger, error),
	) error

	UpdateFundProviderAllocations(
		ctx context.Context,
		walletUUID WalletUUID,
		updateFn func(wallet *Wallet) error,
	) error
}

type WalletUUID struct {
	common.UUID
}

type Wallet struct {
	uuid        WalletUUID
	name        string
	description string

	balance    shared.Money
	currency   shared.Currency
	officeUUID models.OfficeUUID

	fpRegistry *fundProviderRegistry
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

func (w *Wallet) OfficeUUID() models.OfficeUUID {
	return w.officeUUID
}

type WalletAllocationSnapshot struct {
	WalletUUID    WalletUUID
	WalletBalance shared.Money

	AllocationMoney shared.Money

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

	fpBalance, allocationAmount, err := w.fpRegistry.increaseAllocation(fpUUID, amount)
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
		AllocationMoney:     allocationAmount,
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

	fpBalance, allocationAmount, err := w.fpRegistry.decreaseAllocation(fpUUID, amount)
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
		AllocationMoney:     allocationAmount,
		FundProviderUUID:    fpUUID,
		FundProviderBalance: fpBalance,
	}, nil
}

func (w *Wallet) AllocateFundProvider(fp *FundProvider, amount shared.Money) error {
	err := w.fpRegistry.register(fp, amount)
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

func (w *Wallet) FundProviderAllocations() []*FundProviderAllocation {
	allocations := make([]*FundProviderAllocation, 0, len(w.fpRegistry.allocations))

	for _, allocation := range w.fpRegistry.allocations {
		allocations = append(allocations, allocation)
	}

	return allocations
}

func (w *Wallet) FindFundProviderAllocation(fpUUID FundProviderUUID) (*FundProviderAllocation, error) {
	return w.fpRegistry.findAllocation(fpUUID)
}

type NewFundProviderAllocationData struct {
	FundProvider     *FundProvider
	AllocationAmount shared.Money
}

func NewWallet(
	name string,
	description string,
	currency shared.Currency,
	officeUUID models.OfficeUUID,
	allocationsData []NewFundProviderAllocationData,
) (*Wallet, error) {
	errDetails := []common.ErrorDetails{}
	if strings.TrimSpace(name) == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "wallet",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}

	if currency.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "wallet",
			ErrorSlug:  "empty-currency",
			Message:    "currency can't be empty",
		})
	}

	if officeUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "wallet",
			ErrorSlug:  "empty-office-uuid",
			Message:    "office uuid can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-wallet-input", "invalid wallet input").WithDetails(errDetails)
	}

	fpregistry, err := newFundProviderRegistry(currency, officeUUID, allocationsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocations manager: %w", err)
	}

	balance, err := shared.NewMoney(decimal.Zero, currency)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocationsData {
		balance, err = balance.Add(allocation.AllocationAmount)
		if err != nil {
			return nil, err
		}
	}

	return &Wallet{
		uuid:        WalletUUID{UUID: common.NewUUIDv7()},
		name:        name,
		description: description,
		balance:     balance,
		currency:    currency,
		officeUUID:  officeUUID,
		fpRegistry:  fpregistry,
	}, nil
}

type FundProviderAllocation struct {
	fundProvider *FundProvider
	amount       shared.Money
}

func (a *FundProviderAllocation) FundProvider() *FundProvider {
	return a.fundProvider
}

func (a *FundProviderAllocation) Amount() shared.Money {
	return a.amount
}

type fundProviderRegistry struct {
	officeUUID  models.OfficeUUID
	currency    shared.Currency
	allocations map[FundProviderUUID]*FundProviderAllocation
}

func newFundProviderRegistry(currency shared.Currency, officeUUID models.OfficeUUID, allocationsData []NewFundProviderAllocationData) (*fundProviderRegistry, error) {
	allocations := make(map[FundProviderUUID]*FundProviderAllocation, len(allocationsData))

	for _, data := range allocationsData {
		fp := data.FundProvider
		if fp == nil {
			return nil, errors.New("fund provider can't be empty")
		}

		if !fp.officeUUID.UUID.Equals(officeUUID.UUID) {
			return nil, fmt.Errorf("fund provider office (%s) does not match wallet office (%s)",
				fp.OfficeUUID().String(), officeUUID.String())
		}

		if _, ok := allocations[fp.UUID()]; ok {
			return nil, fmt.Errorf("duplicate fund provider '%s' is not allowed", fp.UUID())
		}

		err := data.FundProvider.reserve(data.AllocationAmount)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to allocate fund provider %s with amount %s: %w",
				data.FundProvider.UUID().String(),
				data.AllocationAmount.String(),
				err,
			)
		}

		allocations[data.FundProvider.UUID()] = &FundProviderAllocation{
			fundProvider: data.FundProvider,
			amount:       data.AllocationAmount,
		}
	}

	return &fundProviderRegistry{
		officeUUID:  officeUUID,
		currency:    currency,
		allocations: allocations,
	}, nil
}

func (m *fundProviderRegistry) findAllocation(fpUUID FundProviderUUID) (*FundProviderAllocation, error) {
	allocation, ok := m.allocations[fpUUID]
	if !ok {
		return nil, fmt.Errorf("fund provider %s does not exist", fpUUID.String())
	}

	return allocation, nil
}

func (m *fundProviderRegistry) hasFundProvider(fpUUID FundProviderUUID) bool {
	_, ok := m.allocations[fpUUID]
	return ok
}

func (m *fundProviderRegistry) register(fp *FundProvider, amount shared.Money) error {
	if fp == nil {
		return errors.New("fund provider can't be empty")
	}

	if !m.officeUUID.UUID.Equals(fp.officeUUID.UUID) {
		return errors.New("wallet and fund provider must belong to the same office")
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

	m.allocations[fp.UUID()] = &FundProviderAllocation{
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
