package wallet

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/valueobject"
	"sumni-finance-backend/internal/finance/domain/fundprovider"

	"github.com/google/uuid"
)

type FundProviderAllocationManager struct {
	fpAllocations map[uuid.UUID]*FpAllocation
}

func NewFpAllocationManager(fpAllocations []*FpAllocation) (*FundProviderAllocationManager, error) {
	capacity := len(fpAllocations)
	if capacity == 0 {
		capacity = 1
	}

	allocations := make(map[uuid.UUID]*FpAllocation, capacity)

	for _, allocation := range fpAllocations {
		if allocation == nil || allocation.fp == nil {
			return nil, errors.New("fundProvider can not be nil")
		}

		if allocation.allocated.IsZero() {
			return nil, errors.New("allocated can not be empty")
		}

		_, exist := allocations[allocation.fp.ID()]
		if exist {
			return nil, fmt.Errorf("fundProvider must be unique: %s", allocation.fp.ID())
		}

		allocations[allocation.fp.ID()] = allocation
	}

	return &FundProviderAllocationManager{
		fpAllocations: allocations,
	}, nil
}

func (m *FundProviderAllocationManager) FpAllocations() []FpAllocation {
	fpAllocations := make([]FpAllocation, 0, len(m.fpAllocations))

	for _, allocation := range m.fpAllocations {
		// TODO: consider to check nil
		fpAllocations = append(fpAllocations, *allocation)
	}

	return fpAllocations
}

// FindFundProviderAllocation returns the ProviderAllocation for the given fund provider ID.
// When the bool return value is true, the returned *ProviderAllocation and its provider are guaranteed to be non-nil.
func (m *FundProviderAllocationManager) FindFundProviderAllocation(fpID uuid.UUID) (*FpAllocation, bool) {
	fpAllocation, exist := m.fpAllocations[fpID]

	if !exist || fpAllocation == nil || fpAllocation.fp == nil {
		return nil, false
	}

	return fpAllocation, true
}

func (m *FundProviderAllocationManager) AddFundProviderAndReserve(
	fp *fundprovider.FundProvider,
	allocated valueobject.Money,
) error {
	if fp == nil {
		return errors.New("fund provider is nil")
	}

	if _, exist := m.FindFundProviderAllocation(fp.ID()); exist {
		return ErrFundProviderAlreadyRegistered
	}

	if err := fp.Reserve(allocated); err != nil {
		return err
	}

	fpAllocation, err := NewFpAllocation(fp, allocated.Amount())
	if err != nil {
		return err
	}

	m.fpAllocations[fp.ID()] = fpAllocation
	return nil
}
