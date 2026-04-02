package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(
		ctx context.Context,
		wID uuid.UUID,
	) (*Wallet, error)

	GetByIDWithProviders(
		ctx context.Context,
		wID uuid.UUID,
		spec ProviderAllocationSpec,
	) (*Wallet, error)

	Create(ctx context.Context, wallet *Wallet) error

	CreateAllocations(
		ctx context.Context,
		wID uuid.UUID,
		allocationSpec ProviderAllocationSpec,
		allocatedFunc func(*Wallet) error,
	) error
}
