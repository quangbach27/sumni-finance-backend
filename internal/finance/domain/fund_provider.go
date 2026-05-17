package domain

import (
	"errors"
	"fmt"
	"strings"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

type FundProviderUUID struct {
	common.UUID
}

type FundProviderType struct {
	common.Enum[FundProviderCode]
}

func (at FundProviderType) Code() string {
	return at.String()
}

func MustNewFundProviderType(value string) FundProviderType {
	return common.MustEnum[FundProviderType](value)
}

type FundProviderCode string

func (ts FundProviderCode) Values() []string {
	return []string{"BANK", "CASH"}
}

type FundProviderMetadata interface {
	MatchesType(accountType FundProviderType) bool
	IsZero() bool
}

type BankFundProviderMetadata struct {
	accountNumber string
	bankName      string
}

func NewBankFundProviderMetadata(
	accountNumber string,
	bankName string,
) (BankFundProviderMetadata, error) {
	cleansedAccountNumber := strings.TrimSpace(accountNumber)
	cleansedBankName := strings.TrimSpace(bankName)
	if cleansedAccountNumber == "" {
		return BankFundProviderMetadata{}, errors.New("account number can't be empty")
	}

	if cleansedBankName == "" {
		return BankFundProviderMetadata{}, errors.New("bank name can't be empty")
	}

	return BankFundProviderMetadata{
		accountNumber: cleansedAccountNumber,
		bankName:      cleansedBankName,
	}, nil
}

func (bm BankFundProviderMetadata) AccountNumber() string {
	return bm.accountNumber
}

func (bm BankFundProviderMetadata) BankName() string {
	return bm.bankName
}

func (bm BankFundProviderMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType.String() == "BANK"
}

func (bm BankFundProviderMetadata) IsZero() bool {
	return bm == BankFundProviderMetadata{}
}

type CashFundProviderMetadata struct {
	ownerName string
}

func NewCashFundProviderMetadata(ownerName string) (CashFundProviderMetadata, error) {
	cleansedOwnerName := strings.TrimSpace(ownerName)
	if cleansedOwnerName == "" {
		return CashFundProviderMetadata{}, errors.New("owner name can't be empty")
	}

	return CashFundProviderMetadata{ownerName: cleansedOwnerName}, nil
}

func (cm CashFundProviderMetadata) OwnerName() string {
	return cm.ownerName
}

func (cm CashFundProviderMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType.String() == "CASH"
}

func (cm CashFundProviderMetadata) IsZero() bool {
	return cm == CashFundProviderMetadata{}
}

type FundProvider[T FundProviderMetadata] struct {
	uuid FundProviderUUID

	name             string
	fundProviderType FundProviderType

	balance          shared.Money
	availableBalance shared.Money
	currency         shared.Currency

	metadata T
}

func (fp *FundProvider[T]) UUID() FundProviderUUID {
	return fp.uuid
}

func (fp *FundProvider[T]) Name() string {
	return fp.name
}

func (fp *FundProvider[T]) Type() FundProviderType {
	return fp.fundProviderType
}

func (fp *FundProvider[T]) Balance() shared.Money {
	return fp.balance
}

func (fp *FundProvider[T]) AvailableMoney() shared.Money {
	return fp.availableBalance
}

func (fp *FundProvider[T]) Currency() shared.Currency {
	return fp.currency
}

func (fp *FundProvider[T]) Metadata() T {
	return fp.metadata
}

func (fp *FundProvider[T]) TopUp(m shared.Money) error {
	if m.Amount().IsZero() {
		return errors.New("money for top up can't be empty")
	}

	if m.Amount().IsNegative() {
		return errors.New("money for top up can't be negative")
	}

	newBalance, err := fp.balance.Add(m)
	if err != nil {
		return fmt.Errorf("failed to topup: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundProvider[T]) Withdraw(m shared.Money) error {
	if m.Amount().IsZero() {
		return errors.New("money for withdraw can't be empty")
	}

	if m.Amount().IsNegative() {
		return errors.New("money for withdraw can't be negative")
	}

	newBalance, err := fp.balance.Sub(m)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundProvider[T]) Reserve(m shared.Money) error {
	if !fp.currency.Equal(m.Currency()) {
		return errors.New("allocation money currency must match fund provider currency")
	}

	if m.Amount().IsNegative() {
		return errors.New("allocation money can't be negative")
	}

	if fp.availableBalance.Amount().LessThan(m.Amount()) {
		return errors.New("allocation money can't exceed fund provider unallocated")
	}

	newAvailableBalance, err := fp.availableBalance.Sub(m)
	if err != nil {
		return err
	}

	fp.availableBalance = newAvailableBalance

	return nil
}

func NewFundProvider[T FundProviderMetadata](
	name string,
	fundProviderType FundProviderType,
	initBalance shared.Money,
	metadata T,
) (*FundProvider[T], error) {
	errDetails := []common.ErrorDetails{}

	cleansedName := strings.TrimSpace(name)
	if cleansedName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}

	if fundProviderType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-fund-provider-type",
			Message:    "fund provider type can't be empty",
		})
	}

	if initBalance.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-balance",
			Message:    "balance can't be empty",
		})
	}

	if metadata.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-metadata",
			Message:    "metadata can't not be empty",
		})
	} else {
		if !metadata.MatchesType(fundProviderType) {
			errDetails = append(errDetails, common.ErrorDetails{
				EntityType: "fund_provider",
				ErrorSlug:  "metadata-mismatch",
				Message:    fmt.Sprintf("metadata %T does not match with %s fund provider", metadata, fundProviderType.String()),
			})
		}
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-fund-provider", "invalid fund provider").WithDetails(errDetails)
	}

	return &FundProvider[T]{
		uuid:             FundProviderUUID{UUID: common.NewUUIDv7()},
		name:             cleansedName,
		fundProviderType: fundProviderType,
		balance:          initBalance,
		availableBalance: initBalance,
		currency:         initBalance.Currency(),
		metadata:         metadata,
	}, nil
}
