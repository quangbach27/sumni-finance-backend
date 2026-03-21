package wallet

import (
	"errors"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"

	"github.com/google/uuid"
)

var (
	ErrFundProviderAlreadyRegistered = errors.New("fund provider already registered")
	ErrFundAllocatedMissing          = errors.New("fund provider for allocation is missing")
	ErrAllocationAmountNegative      = errors.New("allocated amount is negative")
)

type Wallet struct {
	id      uuid.UUID
	name    string
	balance valueobject.Money
	version int32

	providerManager *ProviderManager
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
		return ErrFundAllocatedMissing
	}

	if allocatedAmount < 0 {
		return ErrAllocationAmountNegative
	}

	if _, isRegistered := w.ProviderManager().FindProvider(fundProvider.ID()); isRegistered {
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
