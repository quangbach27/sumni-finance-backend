package wallet

import (
	"errors"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
)

type FpAllocation struct {
	fp        *fundprovider.FundProvider
	allocated valueobject.Money
}

func NewFpAllocation(
	fp *fundprovider.FundProvider,
	allocatedAmount int64,
) (*FpAllocation, error) {
	v := validator.New()

	v.Check(fp != nil, "fundProvider", "fundProvider is required")
	v.Check(allocatedAmount >= 0, "allocated", "allocated must be greater or equal 0")

	if err := v.Err(); err != nil {
		return nil, err
	}

	allocated, err := valueobject.NewMoney(allocatedAmount, fp.Currency())
	if err != nil {
		return nil, err
	}

	return &FpAllocation{
		fp:        fp,
		allocated: allocated,
	}, nil
}

func (pa *FpAllocation) FundProvider() *fundprovider.FundProvider { return pa.fp }
func (pa *FpAllocation) Allocated() valueobject.Money             { return pa.allocated }

func (pa *FpAllocation) TopUpFundProviderAndAllocation(amount valueobject.Money) error {
	if err := pa.fp.TopUp(amount); err != nil {
		return err
	}

	newAllocated, err := pa.allocated.Add(amount)
	if err != nil {
		return err
	}

	pa.allocated = newAllocated

	return nil
}

func (pa *FpAllocation) WithdrawFundProviderAndAllocation(amount valueobject.Money) error {
	if amount.GreaterThan(pa.allocated) {
		return errors.New("withdraw amount is excced the allocation amount")
	}

	if err := pa.fp.Withdraw(amount); err != nil {
		return err
	}

	newAllocated, err := pa.allocated.Subtract(amount)
	if err != nil {
		return err
	}

	pa.allocated = newAllocated
	return nil
}
