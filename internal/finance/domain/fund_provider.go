package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/app/models"
)

type FundProviderRepository interface {
	GetFundProviders(ctx context.Context, fundProviderUUID []FundProviderUUID) ([]*FundProvider, error)
	SaveFundProvider(ctx context.Context, fundProvider *FundProvider) error
}

type FundProviderUUID struct {
	common.UUID
}

type FundProviderType struct {
	common.Enum[FundProviderTypeStr]
}

type FundProviderTypeStr string

func (ts FundProviderTypeStr) Values() []string {
	return []string{"bank", "cash"}
}

var (
	FundProviderTypeBank = common.MustEnum[FundProviderType]("bank")
	FundProviderTypeCash = common.MustEnum[FundProviderType]("cash")
)

type FundProviderMetadata interface {
	MatchesType(accountType FundProviderType) bool
	IsZero() bool
}

type FundProviderBankMetadata struct {
	accountNumber string
	bankName      string
	accountOwner  string
}

func NewBankFundProviderMetadata(
	accountNumber string,
	bankName string,
	accountOwner string,
) (FundProviderBankMetadata, error) {
	cleansedAccountNumber := strings.TrimSpace(accountNumber)
	cleansedBankName := strings.TrimSpace(bankName)
	cleansedAccountOwner := strings.TrimSpace(accountOwner)
	if cleansedAccountNumber == "" {
		return FundProviderBankMetadata{}, errors.New("account number can't be empty")
	}

	if cleansedBankName == "" {
		return FundProviderBankMetadata{}, errors.New("bank name can't be empty")
	}

	if cleansedAccountOwner == "" {
		return FundProviderBankMetadata{}, errors.New("bank owner name can't be emtpy")
	}

	return FundProviderBankMetadata{
		accountNumber: cleansedAccountNumber,
		bankName:      cleansedBankName,
		accountOwner:  accountOwner,
	}, nil
}

func (bm FundProviderBankMetadata) AccountNumber() string {
	return bm.accountNumber
}

func (bm FundProviderBankMetadata) BankName() string {
	return bm.bankName
}

func (bm FundProviderBankMetadata) AccountOwner() string {
	return bm.accountOwner
}

func (bm FundProviderBankMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType == FundProviderTypeBank
}

func (bm FundProviderBankMetadata) IsZero() bool {
	return bm == FundProviderBankMetadata{}
}

type FundProviderCashMetadata struct {
	ownerName string
}

func NewCashFundProviderMetadata(ownerName string) (FundProviderCashMetadata, error) {
	cleansedOwnerName := strings.TrimSpace(ownerName)
	if cleansedOwnerName == "" {
		return FundProviderCashMetadata{}, errors.New("owner name can't be empty")
	}

	return FundProviderCashMetadata{ownerName: cleansedOwnerName}, nil
}

func (cm FundProviderCashMetadata) OwnerName() string {
	return cm.ownerName
}

func (cm FundProviderCashMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType == FundProviderTypeCash
}

func (cm FundProviderCashMetadata) IsZero() bool {
	return cm == FundProviderCashMetadata{}
}

type FundProvider struct {
	uuid       FundProviderUUID
	officeUUID models.OfficeUUID

	name             string
	fundProviderType FundProviderType

	balance          shared.Money
	availableBalance shared.Money

	currency shared.Currency

	metadata FundProviderMetadata
}

func (fp *FundProvider) UUID() FundProviderUUID {
	return fp.uuid
}

func (fp *FundProvider) OfficeUUID() models.OfficeUUID {
	return fp.officeUUID
}

func (fp *FundProvider) Name() string {
	return fp.name
}

func (fp *FundProvider) Type() FundProviderType {
	return fp.fundProviderType
}

func (fp *FundProvider) Balance() shared.Money {
	return fp.balance
}

func (fp *FundProvider) AvailableMoney() shared.Money {
	return fp.availableBalance
}

func (fp *FundProvider) Currency() shared.Currency {
	return fp.currency
}

func (fp *FundProvider) Metadata() FundProviderMetadata {
	return fp.metadata
}

func (fp *FundProvider) BankMetadata() (FundProviderBankMetadata, bool) {
	m, ok := fp.metadata.(FundProviderBankMetadata)
	return m, ok
}

func (fp *FundProvider) CashMetadata() (FundProviderCashMetadata, bool) {
	m, ok := fp.metadata.(FundProviderCashMetadata)
	return m, ok
}

func (fp *FundProvider) topUp(m shared.Money) error {
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

func (fp *FundProvider) withdraw(m shared.Money) error {
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

func (fp *FundProvider) reserve(m shared.Money) error {
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

func NewFundProvider(
	name string,
	officeUUID models.OfficeUUID,
	fundProviderType FundProviderType,
	initBalance shared.Money,
	metadata FundProviderMetadata,
) (*FundProvider, error) {
	errDetails := []common.ErrorDetails{}

	cleansedName := strings.TrimSpace(name)
	if cleansedName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}

	if officeUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "fund_provider",
			ErrorSlug:  "empty-office-uuid",
			Message:    "office uuid can't be empty",
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
				Message:    fmt.Sprintf("metadata %T does not match with %s provider", metadata, fundProviderType.String()),
			})
		}
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-fund-provider", "invalid fund provider").WithDetails(errDetails)
	}

	return &FundProvider{
		uuid:             FundProviderUUID{UUID: common.NewUUIDv7()},
		officeUUID:       officeUUID,
		name:             cleansedName,
		fundProviderType: fundProviderType,
		balance:          initBalance,
		availableBalance: initBalance,
		currency:         initBalance.Currency(),
		metadata:         metadata,
	}, nil
}
