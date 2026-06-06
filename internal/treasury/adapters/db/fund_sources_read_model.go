package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/app/query"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FundSourceReadModel struct {
	db *pgxpool.Pool
}

func NewFundSourceReadModel(db *pgxpool.Pool) *FundSourceReadModel {
	if db == nil {
		panic("db can't be nil")
	}
	return &FundSourceReadModel{db: db}
}

func (r *FundSourceReadModel) ListFundSources(ctx context.Context) ([]query.FundSourceView, error) {
	rows, err := dbmodels.New(r.db).ListFundSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing fund sources: %w", err)
	}

	views := make([]query.FundSourceView, len(rows))
	for i, row := range rows {
		views[i] = toFundSourceView(row)
	}
	return views, nil
}

func toFundSourceView(model dbmodels.TreasuryFundSource) query.FundSourceView {
	view := query.FundSourceView{
		UUID:             model.FundSourceUuid,
		Name:             model.Name,
		SourceType:       model.SourceType,
		Balance:          model.Balance,
		AvailableBalance: model.AvailableBalance,
		Currency:         model.Currency,
		CreatedAt:        model.CreatedAt,
		CreatedBy:        model.CreatedBy,
		UpdatedAt:        model.UpdatedAt,
		UpdatedBy:        model.UpdatedBy,
	}

	switch model.SourceType {
	case domain.FundSourceTypeBank:
		view.BankMetadata = &query.FundSourceBankMetadataView{
			BankName:      model.BankInfo.Name(),
			BankCode:      model.BankInfo.BankCode(),
			BankShortName: model.BankInfo.ShortName(),
			BankBin:       model.BankInfo.Bin(),
			BankLogoUrl:   model.BankInfo.LogoUrl(),
			BankIconUrl:   model.BankInfo.IconUrl(),
			AccountNumber: common.SafeDeref(model.BankAccountNumber, ""),
			AccountOwner:  common.SafeDeref(model.BankAccountOwner, ""),
		}
	case domain.FundSourceTypeCash:
		view.CashMetadata = &query.FundSourceCashMetadataView{
			OwnerName: common.SafeDeref(model.CashOwner, ""),
		}
	}

	return view
}
