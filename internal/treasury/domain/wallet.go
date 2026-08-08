package domain

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

type WalletRepository interface {
	CreateWallet(
		ctx context.Context,
		tenantContext shared.TenantContext,
		fundSourceUUIDs []FundSourceUUID,
		createFn func(
			fundSources map[FundSourceUUID]*FundSource,
		) (*Wallet, error),
	) error
}

type WalletUUID struct {
	common.UUID
}

type Wallet struct {
	uuid        WalletUUID
	name        string
	balance     shared.Money
	allocations map[FundSourceUUID]*FundSourceAllocation
}

func (w *Wallet) UUID() WalletUUID {
	return w.uuid
}

func (w *Wallet) Name() string {
	return w.name
}

func (w *Wallet) Balance() shared.Money {
	return w.balance
}

func (w *Wallet) Allocations() map[FundSourceUUID]*FundSourceAllocation {
	return w.allocations
}

func (w *Wallet) Allocate(fs *FundSource, allocatedAmount shared.Money) error {
	if _, exist := w.allocations[fs.UUID()]; exist {
		return fmt.Errorf("fs %s is already allocated", fs.UUID().String())
	}

	err := fs.Reserve(allocatedAmount)
	if err != nil {
		return fmt.Errorf("failed to allocated: %w", err)
	}

	newBalance, err := w.balance.Add(allocatedAmount)
	if err != nil {
		return fmt.Errorf("failed to add allocation for wallet balance: %w", err)
	}
	w.balance = newBalance

	w.allocations[fs.UUID()] = &FundSourceAllocation{
		fundSource: fs,
		balance:    allocatedAmount,
	}
	return nil
}

func (w *Wallet) TopUp(fsUUID FundSourceUUID, amount shared.Money) error {
	if !amount.Amount().IsPositive() {
		return errors.New("top up amount must be positive")
	}

	allocation, exist := w.allocations[fsUUID]
	if !exist {
		return fmt.Errorf("fs %s does not in wallet allocation %s", fsUUID.String(), w.UUID().String())
	}

	err := allocation.fundSource.topUp(amount)
	if err != nil {
		return fmt.Errorf("failed to top up fund source: %w", err)
	}
	newAllocationBalance, err := allocation.balance.Add(amount)
	if err != nil {
		return fmt.Errorf("failed to topup allocation balance: %w", err)
	}
	allocation.balance = newAllocationBalance

	newWalletBalance, err := w.balance.Add(amount)
	if err != nil {
		return fmt.Errorf("failed to topup wallet balance: %w", err)
	}
	w.balance = newWalletBalance

	return nil
}

func (w *Wallet) Withdraw(fsUUID FundSourceUUID, amount shared.Money) error {
	if !amount.Amount().IsPositive() {
		return errors.New("withdraw amount must be positive")
	}

	allocation, exist := w.allocations[fsUUID]
	if !exist {
		return fmt.Errorf("fs %s does not in wallet allocation %s", fsUUID.String(), w.UUID().String())
	}

	if !allocation.balance.Amount().GreaterThanOrEqual(amount.Amount()) {
		return fmt.Errorf("wallet doesn't have enough money")
	}

	err := allocation.fundSource.withdraw(amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw fund source: %w", err)
	}

	newAllocationBalance, err := allocation.balance.Sub(amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw allocation money: %w", err)
	}
	allocation.balance = newAllocationBalance

	newWalletBalance, err := w.balance.Sub(amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw wallet balance: %w", err)
	}
	w.balance = newWalletBalance

	return nil
}

func NewWallet(
	name string,
	currency shared.Currency,
) (*Wallet, error) {
	errDetails := []common.ErrorDetails{}
	if name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Wallet",
			ErrorSlug:  "name-empty",
			Message:    "name can't be empty",
		})
	}

	if currency.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Wallet",
			ErrorSlug:  "currency-empty",
			Message:    "currency can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-wallet", "wallet is invalid").WithDetails(errDetails)
	}

	initBalance, err := shared.NewMoney(decimal.Zero, currency)
	if err != nil {
		return nil, err
	}
	return &Wallet{
		uuid:        WalletUUID{common.NewUUIDv7()},
		name:        name,
		balance:     initBalance,
		allocations: make(map[FundSourceUUID]*FundSourceAllocation),
	}, nil
}

type FundSourceAllocation struct {
	fundSource *FundSource
	balance    shared.Money
}

func (f *FundSourceAllocation) FundSource() *FundSource {
	return f.fundSource
}

func (f *FundSourceAllocation) Balance() shared.Money {
	return f.balance
}
