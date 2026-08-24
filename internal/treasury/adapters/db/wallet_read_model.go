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

type walletReadModel struct {
	db *pgxpool.Pool
}

var _ query.WalletReadModel = (*walletReadModel)(nil)

func NewWalletReadModel(db *pgxpool.Pool) *walletReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	return &walletReadModel{db: db}
}

func (r *walletReadModel) ListWallets(ctx context.Context, q query.ListWallets) (query.ListWalletsReadModel, error) {
	queries := dbmodels.New(r.db)

	offset := int32((q.Page - 1) * q.PageSize)
	limit := int32(q.PageSize)

	wallets, err := queries.ListWallets(ctx, dbmodels.ListWalletsParams{
		TenantID:  q.TenantContext.TenantID(),
		OfficeID:  q.TenantContext.OfficeID(),
		OffsetVal: offset,
		LimitVal:  limit,
	})
	if err != nil {
		return query.ListWalletsReadModel{}, fmt.Errorf("error listing wallets: %w", err)
	}

	totalCount, err := queries.CountWallets(ctx, dbmodels.CountWalletsParams{
		TenantID: q.TenantContext.TenantID(),
		OfficeID: q.TenantContext.OfficeID(),
	})
	if err != nil {
		return query.ListWalletsReadModel{}, fmt.Errorf("error counting wallets: %w", err)
	}

	walletUUIDs := make([]common.UUID, len(wallets))
	for i, w := range wallets {
		walletUUIDs[i] = w.WalletUuid.UUID
	}

	allocationRows, err := queries.GetAllocationsByWalletUUIDs(ctx, walletUUIDs)
	if err != nil {
		return query.ListWalletsReadModel{}, fmt.Errorf("error loading wallet allocations: %w", err)
	}

	allocationsByWallet := make(map[domain.WalletUUID][]query.WalletAllocationItem)
	for _, row := range allocationRows {
		wID := row.TreasuryWalletAllocation.WalletUuid
		allocationsByWallet[wID] = append(allocationsByWallet[wID], query.WalletAllocationItem{
			FundSourceUUID: row.TreasuryFundSource.FundSourceUuid,
			FundSourceName: row.TreasuryFundSource.Name,
			Balance:        row.TreasuryWalletAllocation.Balance,
		})
	}

	items := make([]query.ListWalletItem, 0, len(wallets))
	for _, w := range wallets {
		items = append(items, query.ListWalletItem{
			UUID:        w.WalletUuid,
			Name:        w.Name,
			Currency:    w.Currency,
			Balance:     w.Balance,
			Allocations: allocationsByWallet[w.WalletUuid],
		})
	}

	return query.ListWalletsReadModel(common.Pagination[query.ListWalletItem]{
		Items:      items,
		TotalCount: int(totalCount),
		Page:       q.Page,
		PageSize:   q.PageSize,
	}), nil
}
