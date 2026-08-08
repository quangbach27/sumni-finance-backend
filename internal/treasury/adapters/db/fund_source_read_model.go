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

type fundSourceReadModel struct {
	db *pgxpool.Pool
}

func NewFundSourceReadModel(db *pgxpool.Pool) *fundSourceReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	return &fundSourceReadModel{db: db}
}

func (r *fundSourceReadModel) ListFundSources(ctx context.Context, q query.ListFundSources) (query.ListFundSourcesReadModel, error) {
	queries := dbmodels.New(r.db)

	offset := int32((q.Page - 1) * q.PageSize)
	limit := int32(q.PageSize)

	rows, err := queries.ListFundSources(ctx, dbmodels.ListFundSourcesParams{
		TenantID:  q.TenantContext.TenantID(),
		OfficeID:  q.TenantContext.OfficeID(),
		OffsetVal: offset,
		LimitVal:  limit,
	})
	if err != nil {
		return query.ListFundSourcesReadModel{}, fmt.Errorf("error listing fund sources: %w", err)
	}

	totalCount, err := queries.CountFundSources(ctx, dbmodels.CountFundSourcesParams{
		TenantID: q.TenantContext.TenantID(),
		OfficeID: q.TenantContext.OfficeID(),
	})
	if err != nil {
		return query.ListFundSourcesReadModel{}, fmt.Errorf("error counting fund sources: %w", err)
	}

	items := make([]query.ListFundSourceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toFundSourceItem(row))
	}

	return query.ListFundSourcesReadModel(common.Pagination[query.ListFundSourceItem]{
		Items:      items,
		TotalCount: int(totalCount),
		Page:       q.Page,
		PageSize:   q.PageSize,
	}), nil
}

func toFundSourceItem(row dbmodels.TreasuryFundSource) query.ListFundSourceItem {
	item := query.ListFundSourceItem{
		UUID:             row.FundSourceUuid,
		Name:             row.Name,
		SourceType:       row.SourceType,
		Currency:         row.Currency,
		Balance:          row.Balance,
		AvailableBalance: row.AvailableBalance,
	}

	if row.SourceType == domain.FundSourceTypeBank {
		item.BankMetadata = &query.FundSourceBankMetadata{
			BankName:      row.BankInfo.Name(),
			BankBin:       row.BankInfo.Bin(),
			BankShortName: row.BankInfo.ShortName(),
			AccountNumber: common.SafeDeref(row.BankAccountNumber, ""),
			AccountOwner:  common.SafeDeref(row.BankAccountOwner, ""),
		}
	}

	if row.SourceType == domain.FundSourceTypeCash {
		item.CashMetadata = &query.FundSourceCashMetadata{
			OwnerName: common.SafeDeref(row.CashOwner, ""),
		}
	}

	return item
}
