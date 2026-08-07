package domain

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

type FundSourceRepository interface {
	CreateFundSource(ctx context.Context, tenant shared.TenantContext, fundSource *FundSource) error
}

type FundSourceUUID struct {
	common.UUID
}

type FundSourceType struct {
	common.Enum[FundSourceTypeValues]
}

type FundSourceTypeValues string

func (ts FundSourceTypeValues) Values() []string {
	return []string{"BANK", "CASH"}
}

var (
	FundSourceTypeBank = common.MustEnum[FundSourceType]("BANK")
	FundSourceTypeCash = common.MustEnum[FundSourceType]("CASH")
)

type FundSource struct {
	uuid             FundSourceUUID
	name             string
	sourceType       FundSourceType
	balance          shared.Money
	availableBalance shared.Money
	currency         shared.Currency

	metadata FundSourceMetadata
}

func NewFundSource(
	name string,
	sourceType FundSourceType,
	initBalance shared.Money,
	currency shared.Currency,
	metadata FundSourceMetadata,
) (*FundSource, error) {
	errDetails := []common.ErrorDetails{}

	if name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}
	if sourceType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-source-type",
			Message:    "source type can't be empty",
		})
	}
	if currency.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-currency",
			Message:    "currency can't be empty",
		})
	}
	if initBalance.Amount().IsNegative() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "negative-balance",
			Message:    "balance can't be negative",
		})
	}
	if metadata == nil || metadata.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-metadata",
			Message:    "metadata can't be empty",
		})
	} else if !metadata.MatchesType(sourceType) {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "metadata-type-mismatch",
			Message:    "metadata type does not match source type",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-fund-source-metadata", "fund source metadata is not valid").WithDetails(errDetails)
	}
	return &FundSource{
		uuid:             FundSourceUUID{UUID: common.NewUUIDv7()},
		name:             name,
		sourceType:       sourceType,
		balance:          initBalance,
		availableBalance: initBalance,
		currency:         currency,
		metadata:         metadata,
	}, nil
}

func (f *FundSource) UUID() FundSourceUUID           { return f.uuid }
func (f *FundSource) Name() string                   { return f.name }
func (f *FundSource) SourceType() FundSourceType     { return f.sourceType }
func (f *FundSource) Balance() shared.Money          { return f.balance }
func (f *FundSource) AvailableBalance() shared.Money { return f.availableBalance }
func (f *FundSource) Currency() shared.Currency      { return f.currency }
func (f *FundSource) Metadata() FundSourceMetadata   { return f.metadata }
func (fp *FundSource) BankMetadata() (BankMetadata, bool) {
	m, ok := fp.metadata.(BankMetadata)
	return m, ok
}

func (fp *FundSource) CashMetadata() (CashMetadata, bool) {
	m, ok := fp.metadata.(CashMetadata)
	return m, ok
}

func (fp *FundSource) topUp(m shared.Money) error {
	if !m.Amount().IsPositive() {
		return errors.New("top up amount must be positive")
	}

	newBalance, err := fp.balance.Add(m)
	if err != nil {
		return fmt.Errorf("failed to topup: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundSource) withdraw(m shared.Money) error {
	if !m.Amount().IsPositive() {
		return errors.New("withdraw amount must be positive")
	}

	if fp.balance.Amount().LessThan(m.Amount()) {
		return errors.New("insufficient balance to complete withdrawal")
	}

	newBalance, err := fp.balance.Sub(m)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundSource) Reserve(m shared.Money) error {
	if !fp.currency.Equal(m.Currency()) {
		return fmt.Errorf("reservation currency %s does not match fund source currency %s", m.Currency(), fp.currency)
	}

	if m.Amount().IsNegative() {
		return errors.New("reservation amount must be greater or equal than zero")
	}

	if fp.availableBalance.Amount().LessThan(m.Amount()) {
		return errors.New("insufficient available balance to complete reservation")
	}

	newAvailableBalance, err := fp.availableBalance.Sub(m)
	if err != nil {
		return err
	}

	fp.availableBalance = newAvailableBalance

	return nil
}
