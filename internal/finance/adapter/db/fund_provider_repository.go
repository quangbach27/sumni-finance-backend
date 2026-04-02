package db

import (
	"context"
	"errors"
	"sumni-finance-backend/internal/finance/adapter/db/store"
	"sumni-finance-backend/internal/finance/domain/fundprovider"

	"github.com/google/uuid"
)

type fundProviderRepo struct {
	queries *store.Queries
}

func NewFundProviderRepo(
	queries *store.Queries,
) (*fundProviderRepo, error) {
	if queries == nil {
		return nil, errors.New("missing dependencies")
	}

	return &fundProviderRepo{
		queries: queries,
	}, nil
}

func (r *fundProviderRepo) Create(
	ctx context.Context,
	fp *fundprovider.FundProvider,
) error {
	return r.queries.CreateFundProvider(ctx, store.CreateFundProviderParams{
		ID:                fp.ID(),
		Name:              fp.Name(),
		FpType:            fp.Type().String(),
		Balance:           fp.Balance().Amount(),
		Currency:          fp.Currency().Code(),
		UnallocatedAmount: fp.UnallocatedBalance().Amount(),
		Version:           fp.Version(),
	})
}

func (r *fundProviderRepo) GetByID(ctx context.Context, fpID uuid.UUID) (*fundprovider.FundProvider, error) {
	fpModel, err := r.queries.GetFundProviderByID(ctx, fpID)
	if err != nil {
		return nil, err
	}

	return fundprovider.UnmarshalFundProviderFromDatabase(
		fpModel.ID,
		fpModel.Name,
		fpModel.FpType,
		fpModel.Balance,
		fpModel.UnallocatedAmount,
		fpModel.Currency,
		fpModel.Version,
	)
}

func (r *fundProviderRepo) GetByIDs(ctx context.Context, fpID []uuid.UUID) ([]*fundprovider.FundProvider, error) {
	fpModels, err := r.queries.GetFundProvidersByIDs(ctx, fpID)
	if err != nil {
		return nil, err
	}

	fps := make([]*fundprovider.FundProvider, 0, len(fpModels))
	for _, fpModel := range fpModels {
		fp, err := fundprovider.UnmarshalFundProviderFromDatabase(
			fpModel.ID,
			fpModel.Name,
			fpModel.FpType,
			fpModel.Balance,
			fpModel.UnallocatedAmount,
			fpModel.Currency,
			fpModel.Version,
		)
		if err != nil {
			return nil, err
		}
		fps = append(fps, fp)
	}

	return fps, nil
}
