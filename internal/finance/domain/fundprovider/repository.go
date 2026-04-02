package fundprovider

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, fp *FundProvider) error
	GetByID(ctx context.Context, fpID uuid.UUID) (*FundProvider, error)
	GetByIDs(ctx context.Context, fpID []uuid.UUID) ([]*FundProvider, error)
}
