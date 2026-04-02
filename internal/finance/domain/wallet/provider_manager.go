package wallet

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"

	"github.com/google/uuid"
)

type ProviderManager struct {
	providers map[uuid.UUID]ProviderAllocation
}

func NewProviderManager(allocations []ProviderAllocation) (*ProviderManager, error) {
	providers := make(map[uuid.UUID]ProviderAllocation, len(allocations))

	for _, allocation := range allocations {
		if allocation.provider == nil {
			return nil, errors.New("fundProvider can not be nil")
		}

		if allocation.allocated.IsZero() {
			return nil, errors.New("allocated is required")
		}

		_, exist := providers[allocation.provider.ID()]
		if exist {
			return nil, fmt.Errorf("fundProvider must be unique: %s", allocation.provider.ID())
		}

		providers[allocation.provider.ID()] = allocation
	}

	return &ProviderManager{
		providers: providers,
	}, nil
}

func (m *ProviderManager) ProviderAllocations() []ProviderAllocation {
	providerAllocations := make([]ProviderAllocation, 0, len(m.providers))

	for _, provider := range m.providers {
		providerAllocations = append(providerAllocations, provider)
	}

	return providerAllocations
}

func (m *ProviderManager) FindProvider(id uuid.UUID) *fundprovider.FundProvider {
	allocation := m.providers[id]
	return allocation.provider
}

func (m *ProviderManager) AddFundProviderAndReserve(
	fundProvider *fundprovider.FundProvider,
	allocated valueobject.Money,
) error {
	if err := fundProvider.Reserve(allocated); err != nil {
		return err
	}

	providerAllocation, err := NewProviderAllocation(fundProvider, allocated.Amount())
	if err != nil {
		return err
	}

	m.providers[fundProvider.ID()] = providerAllocation
	return nil
}
