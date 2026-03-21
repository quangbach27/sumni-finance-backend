package fundprovider

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"

	"github.com/google/uuid"
)

var (
	ErrInsufficientAmount = errors.New("amount must be greater or equal 0")
)

type ErrInsufficientAllocatedAmount struct {
	AllocatedAmount   int64
	UnallocatedAmount int64
}

func (err ErrInsufficientAllocatedAmount) Error() string {
	return fmt.Sprintf("allocated amount '%d' has exccedd unallocated amount '%d' of fund provider", err.AllocatedAmount, err.UnallocatedAmount)
}

type ErrInsufficientWithdrawAmount struct {
	WithdrawAmount  int64
	AllocatedAmount int64
}

func (err ErrInsufficientWithdrawAmount) Error() string {
	return fmt.Sprintf("withdraw amount '%d' has excceedd allocated amount '%d' of fund provider", err.WithdrawAmount, err.AllocatedAmount)
}

type FundProvider struct {
	id                 uuid.UUID
	name               string
	fpType             Type
	balance            valueobject.Money
	unallocatedBalance valueobject.Money
	version            int32
}

func NewFundProvider(
	name string,
	typeStr string,
	initBalanceAmount int64,
	currencyCode string,
) (*FundProvider, error) {
	v := validator.New()

	v.Required(name, "name")
	v.Check(initBalanceAmount >= 0, "initBalance", "initBalance must be greater or equal than 0")
	v.Required(currencyCode, "currency")

	if err := v.Err(); err != nil {
		return nil, err
	}

	fpType, err := NewType(typeStr)
	if err != nil {
		return nil, err
	}

	currency, err := valueobject.NewCurrency(currencyCode)
	if err != nil {
		return nil, err
	}

	initBalance, err := valueobject.NewMoney(initBalanceAmount, currency)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to create fundProviderID: %w", err)
	}

	return &FundProvider{
		id:                 id,
		name:               name,
		fpType:             fpType,
		balance:            initBalance,
		unallocatedBalance: initBalance,
		version:            0,
	}, nil
}

func UnmarshalFundProviderFromDatabase(
	id uuid.UUID,
	name string,
	typeStr string,
	balanceAmount int64,
	unallocatedBalanceAmount int64,
	currencyCode string,
	version int32,
) (*FundProvider, error) {
	v := validator.New()

	v.Check(id != uuid.Nil, "id", "id is required")
	v.Required(name, "name")
	v.Check(balanceAmount >= 0, "balance", "balance must greater or equal than 0")
	v.Check(unallocatedBalanceAmount >= 0, "unallocatedBalance", "unallocatedBalance must greater or equal than 0")
	v.Check(balanceAmount >= unallocatedBalanceAmount, "unallocatedBalanceAmount", "unallocatedBalanceAmount must smaller than provider balance")
	v.Check(version >= 0, "version", "version must greater or equal than 0")

	if err := v.Err(); err != nil {
		return nil, err
	}

	fpType, err := NewType(typeStr)
	if err != nil {
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

	unallocatedBalance, err := valueobject.NewMoney(unallocatedBalanceAmount, currency)
	if err != nil {
		return nil, err
	}

	return &FundProvider{
		id:                 id,
		name:               name,
		fpType:             fpType,
		balance:            balance,
		unallocatedBalance: unallocatedBalance,
		version:            version,
	}, nil
}

func (p *FundProvider) ID() uuid.UUID                         { return p.id }
func (p *FundProvider) Name() string                          { return p.name }
func (p *FundProvider) Type() Type                            { return p.fpType }
func (p *FundProvider) Balance() valueobject.Money            { return p.balance }
func (p *FundProvider) Currency() valueobject.Currency        { return p.balance.Currency() }
func (p *FundProvider) UnallocatedBalance() valueobject.Money { return p.unallocatedBalance }
func (p *FundProvider) Version() int32                        { return p.version }
func (p *FundProvider) AllocatedBalance() valueobject.Money {
	allocatedBalance, _ := p.balance.Subtract(p.unallocatedBalance)
	return allocatedBalance
}

func (p *FundProvider) TopUp(amount valueobject.Money) error {
	if amount.IsNegative() || amount.Amount() == 0 {
		return ErrInsufficientAmount
	}

	newBalance, err := p.balance.Add(amount)
	if err != nil {
		return err
	}

	p.balance = newBalance
	return nil
}

func (p *FundProvider) Withdraw(amount valueobject.Money) error {
	if amount.IsNegative() || amount.Amount() == 0 {
		return ErrInsufficientAmount
	}

	if amount.GreaterThan(p.AllocatedBalance()) {
		return ErrInsufficientWithdrawAmount{
			WithdrawAmount:  amount.Amount(),
			AllocatedAmount: p.AllocatedBalance().Amount(),
		}
	}

	newBalance, err := p.balance.Subtract(amount)
	if err != nil {
		return err
	}

	p.balance = newBalance
	return nil
}

// Allocate reserves a portion of the provider's available funds for a wallet.
// It reduces the availableAmountForAllocation by the specified allocatedAmount.
// Returns ErrInsufficientAvailable if the requested amount exceeds the available balance.
func (p *FundProvider) Reserve(
	allocated valueobject.Money,
) error {
	if allocated.IsNegative() {
		return ErrInsufficientAmount
	}

	if allocated.GreaterThan(p.unallocatedBalance) {
		return ErrInsufficientAllocatedAmount{
			AllocatedAmount:   allocated.Amount(),
			UnallocatedAmount: p.unallocatedBalance.Amount(),
		}
	}

	newUnallocatedAmount, err := p.unallocatedBalance.Subtract(allocated)
	if err != nil {
		return err
	}

	p.unallocatedBalance = newUnallocatedAmount
	return nil
}
