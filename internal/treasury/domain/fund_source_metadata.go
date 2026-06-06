package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"sumni-finance-backend/internal/common"
)

type FundSourceMetadata interface {
	MatchesType(accountType FundSourceType) bool
	IsZero() bool
}

type FundSourceBankMetadata struct {
	bankInfo      BankInfo
	accountNumber string
	accountOwner  string
}

func (bm FundSourceBankMetadata) BankInfo() BankInfo {
	return bm.bankInfo
}

func (bm FundSourceBankMetadata) AccountNumber() string {
	return bm.accountNumber
}

func (bm FundSourceBankMetadata) AccountOwner() string {
	return bm.accountOwner
}

func (bm FundSourceBankMetadata) MatchesType(sourceType FundSourceType) bool {
	return sourceType == FundSourceTypeBank
}

func (bm FundSourceBankMetadata) IsZero() bool {
	return bm == FundSourceBankMetadata{}
}

type FundSourceCashMetadata struct {
	ownerName string
}

func (cm FundSourceCashMetadata) OwnerName() string {
	return cm.ownerName
}

func (cm FundSourceCashMetadata) MatchesType(sourceType FundSourceType) bool {
	return sourceType == FundSourceTypeCash
}

func (cm FundSourceCashMetadata) IsZero() bool {
	return cm == FundSourceCashMetadata{}
}

type BankInfo struct {
	name            string
	bin             int
	bankCode        string
	shortName       string
	logoUrl         string
	iconUrl         string
	lookupSupport   int
	transferSupport int
}

func NewBankInfo(
	name string,
	bin int,
	bankCode string,
	shortName string,
	logoUrl string,
	iconUrl string,
	lookupSupport int,
	transferSupport int,
) (BankInfo, error) {
	errDetails := []common.ErrorDetails{}

	if name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "name",
			Message:    "name is required",
		})
	}
	if bin == 0 {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "bin",
			Message:    "bin is required",
		})
	}
	if bankCode == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "bankCode",
			Message:    "bank code is required",
		})
	}
	if shortName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "shortName",
			Message:    "short name is required",
		})
	}
	if logoUrl == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "logoUrl",
			Message:    "logo url is required",
		})
	}

	if iconUrl == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankInfo",
			ErrorSlug:  "iconUrl",
			Message:    "icon url is required",
		})
	}

	if len(errDetails) > 0 {
		return BankInfo{}, common.NewInvalidInputError("invalid-bank-info", "bank info is not valid").WithDetails(errDetails)
	}

	return BankInfo{
		name:            name,
		bin:             bin,
		bankCode:        bankCode,
		shortName:       shortName,
		logoUrl:         logoUrl,
		iconUrl:         iconUrl,
		lookupSupport:   lookupSupport,
		transferSupport: transferSupport,
	}, nil
}

func (b BankInfo) Name() string         { return b.name }
func (b BankInfo) Bin() int             { return b.bin }
func (b BankInfo) BankCode() string     { return b.bankCode }
func (b BankInfo) ShortName() string    { return b.shortName }
func (b BankInfo) LogoUrl() string      { return b.logoUrl }
func (b BankInfo) IconUrl() string      { return b.iconUrl }
func (b BankInfo) LookupSupport() int   { return b.lookupSupport }
func (b BankInfo) TransferSupport() int { return b.transferSupport }
func (b BankInfo) IsZero() bool         { return b == BankInfo{} }

type bankInfoDto struct {
	Name            string `json:"name"`
	Bin             int    `json:"bin"`
	BankCode        string `json:"bankCode"`
	ShortName       string `json:"shortName"`
	LogoUrl         string `json:"logoUrl"`
	IconUrl         string `json:"iconUrl"`
	LookupSupport   int    `json:"lookupSupport"`
	TransferSupport int    `json:"transferSupport"`
}

func (b BankInfo) Value() (driver.Value, error) {
	dto := bankInfoDto{
		Name:            b.name,
		Bin:             b.bin,
		BankCode:        b.bankCode,
		ShortName:       b.shortName,
		LogoUrl:         b.logoUrl,
		IconUrl:         b.iconUrl,
		LookupSupport:   b.lookupSupport,
		TransferSupport: b.transferSupport,
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
	b.bankCode = dto.BankCode
	b.shortName = dto.ShortName
	b.logoUrl = dto.LogoUrl
	b.iconUrl = dto.IconUrl
	b.lookupSupport = dto.LookupSupport
	b.transferSupport = dto.TransferSupport
	return nil
}
