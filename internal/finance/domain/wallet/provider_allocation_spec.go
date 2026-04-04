package wallet

import "github.com/google/uuid"

type ProviderAllocationSpec interface {
	IsSatisfiedBy(p FpAllocation) bool
}

type DefaultProviderAllocationSpec struct{}

func NewDefaultProviderAllocationSpec() DefaultProviderAllocationSpec {
	return DefaultProviderAllocationSpec{}
}

func (spec DefaultProviderAllocationSpec) IsSatisfiedBy(p FpAllocation) bool {
	return true
}

type ProviderMatchesAnySpec struct {
	allowed map[uuid.UUID]struct{}
}

func NewProviderMatchesAnySpec(providerIDs []uuid.UUID) ProviderMatchesAnySpec {
	allowed := make(map[uuid.UUID]struct{})

	for _, id := range providerIDs {
		allowed[id] = struct{}{}
	}

	return ProviderMatchesAnySpec{
		allowed: allowed,
	}
}

func (spec ProviderMatchesAnySpec) IsSatisfiedBy(p FpAllocation) bool {
	provider := p.FundProvider()
	if provider == nil {
		return false
	}

	_, exists := spec.allowed[provider.ID()]
	return exists
}
