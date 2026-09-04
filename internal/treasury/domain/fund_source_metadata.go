package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"sumni-finance-backend/internal/common"
)

type FundSourceMetadata interface {
	MatchesType(fsType FundSourceType) bool
	IsZero() bool
}

type BankMetadata struct {
	bankInfo      BankInfo
	accountNumber string
	accountOwner  string
}

type BankInfoData struct {
	Name      string
	Bin       string
	ShortName string
}

func NewBankMetadata(
	accountNumber string,
	accountOwner string,
	bankInfoData BankInfoData,
) (BankMetadata, error) {
	errDetails := []common.ErrorDetails{}
	if accountNumber == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceBankMetadata",
			ErrorSlug:  "empty-account-number",
			Message:    "account number can't be empty",
		})
	}
	if accountOwner == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceBankMetadata",
			ErrorSlug:  "empty-account-owner",
			Message:    "account owner can't be empty",
		})
	}
	if bankInfoData.Name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "name",
			Message:    "name is required",
		})
	}
	if bankInfoData.Bin == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "bin",
			Message:    "bin is required",
		})
	}
	if bankInfoData.ShortName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "shortName",
			Message:    "short name is required",
		})
	}
	if len(errDetails) != 0 {
		return BankMetadata{}, common.NewInvalidInputError("invalid-bank-metadata", "bank metadata is not valid").WithDetails(errDetails)
	}

	return BankMetadata{
		bankInfo: BankInfo{
			name:      bankInfoData.Name,
			bin:       bankInfoData.Bin,
			shortName: bankInfoData.ShortName,
		},
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
	}, nil
}

func (bm BankMetadata) BankInfo() BankInfo {
	return bm.bankInfo
}

func (bm BankMetadata) AccountNumber() string {
	return bm.accountNumber
}

func (bm BankMetadata) AccountOwner() string {
	return bm.accountOwner
}

func (bm BankMetadata) MatchesType(fsType FundSourceType) bool {
	return fsType == FundSourceTypeBank
}

func (bm BankMetadata) IsZero() bool {
	return bm == BankMetadata{}
}

type BankInfo struct {
	name      string
	bin       string
	shortName string
}

func (b BankInfo) Name() string      { return b.name }
func (b BankInfo) Bin() string       { return b.bin }
func (b BankInfo) ShortName() string { return b.shortName }
func (b BankInfo) IsZero() bool      { return b == BankInfo{} }

type bankInfoDto struct {
	Name      string `json:"name"`
	Bin       string `json:"bin"`
	ShortName string `json:"short_name"`
}

func (b BankInfo) Value() (driver.Value, error) {
	dto := bankInfoDto{
		Name:      b.name,
		Bin:       b.bin,
		ShortName: b.shortName,
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BankInfo: %w", err)
	}
	return string(data), nil
}

func (b *BankInfo) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for BankInfo: %T", src)
	}

	var dto bankInfoDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return fmt.Errorf("failed to unmarshal BankInfo: %w", err)
	}

	b.name = dto.Name
	b.bin = dto.Bin
	b.shortName = dto.ShortName
	return nil
}

type CashMetadata struct {
	ownerName string
}

func (cm CashMetadata) OwnerName() string {
	return cm.ownerName
}

func (cm CashMetadata) MatchesType(fsType FundSourceType) bool {
	return fsType == FundSourceTypeCash
}

func (cm CashMetadata) IsZero() bool {
	return cm == CashMetadata{}
}

func NewCashMetadata(ownerName string) (CashMetadata, error) {
	errDetails := []common.ErrorDetails{}
	if ownerName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceCashMetadata",
			ErrorSlug:  "empty-owner-name",
			Message:    "owner name can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return CashMetadata{}, common.NewInvalidInputError("invalid-cash-metadata", "cash metadata is not valid").WithDetails(errDetails)
	}

	return CashMetadata{ownerName: ownerName}, nil
}
