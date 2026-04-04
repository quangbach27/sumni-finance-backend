package command_test

import (
	"context"
	"sumni-finance-backend/internal/finance/app/command"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	fp_mocks "sumni-finance-backend/internal/finance/domain/fundprovider/mocks"
	"sumni-finance-backend/internal/finance/domain/wallet"
	wallet_mocks "sumni-finance-backend/internal/finance/domain/wallet/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type AllocateFundDependenciesManager struct {
	fundProviderRepoMock *fp_mocks.MockRepository
	walletRepoMock       *wallet_mocks.MockRepository
}

func NewAllocateFundDM(t *testing.T) *AllocateFundDependenciesManager {
	t.Helper()

	return &AllocateFundDependenciesManager{
		fundProviderRepoMock: fp_mocks.NewMockRepository(t),
		walletRepoMock:       wallet_mocks.NewMockRepository(t),
	}
}

func (dm *AllocateFundDependenciesManager) NewHandler() command.AllocateFundHandler {
	return command.NewAllocateFundHandler(dm.walletRepoMock, dm.fundProviderRepoMock)
}

func TestAllocateFundHandler_Handle(t *testing.T) {
	t.Run("returns errors when providers cmd is empty", func(t *testing.T) {
		cmd := command.AllocateFundCmd{}

		err := NewAllocateFundDM(t).NewHandler().Handle(context.Background(), cmd)
		require.Error(t, err)
	})

	t.Run("returns errors when provider repo getByIDs fails", func(t *testing.T) {
		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: uuid.New(), AllocatedAmount: 50},
				{ID: uuid.New(), AllocatedAmount: 50},
			},
		}

		dm := NewAllocateFundDM(t)
		dm.fundProviderRepoMock.EXPECT().GetByIDs(mock.Anything, mock.Anything).Return(nil, assert.AnError)

		err := dm.NewHandler().Handle(context.Background(), cmd)

		require.Error(t, err)
	})

	t.Run("returns error when repository returns fewer providers than requested", func(t *testing.T) {
		provider1, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		provider2, err := fundprovider.NewFundProvider("TPBank", "BANK", 100, "USD")
		require.NoError(t, err)

		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: provider1.ID(), AllocatedAmount: 50},
				{ID: provider2.ID(), AllocatedAmount: 50},
			},
		}

		dm := NewAllocateFundDM(t)
		dm.fundProviderRepoMock.
			EXPECT().
			GetByIDs(mock.Anything, mock.Anything).
			Return(
				[]*fundprovider.FundProvider{
					provider1,
				},
				nil,
			)

		err = dm.NewHandler().Handle(context.Background(), cmd)

		require.Error(t, err)
	})

	t.Run("return error when allocated amount excceed unallocated of fund provider", func(t *testing.T) {
		provider1, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: provider1.ID(), AllocatedAmount: 110},
			},
		}

		dm := NewAllocateFundDM(t)

		// 1. Mock GetByIDs
		dm.fundProviderRepoMock.
			EXPECT().
			GetByIDs(mock.Anything, mock.Anything).
			Return(
				[]*fundprovider.FundProvider{
					provider1,
				},
				nil,
			).
			Once()

		dm.walletRepoMock.
			EXPECT().
			CreateAllocations(
				mock.Anything,
				cmd.WalletID,
				mock.Anything,
				mock.Anything,
			).
			RunAndReturn(func(
				ctx context.Context,
				wID uuid.UUID,
				spec wallet.ProviderAllocationSpec,
				updateFunc func(*wallet.Wallet) error,
			) error {
				w, err := wallet.NewWallet("USD", "Tai chinh tong")
				require.NoError(t, err)

				return updateFunc(w)
			})

		err = dm.NewHandler().Handle(context.Background(), cmd)
		require.Error(t, err)
	})

	t.Run("returns error when provider is already allocated", func(t *testing.T) {
		provider1, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: provider1.ID(), AllocatedAmount: 50},
			},
		}

		dm := NewAllocateFundDM(t)

		// 1. Mock GetByIDs
		dm.fundProviderRepoMock.
			EXPECT().
			GetByIDs(mock.Anything, mock.Anything).
			Return(
				[]*fundprovider.FundProvider{
					provider1,
				},
				nil,
			).
			Once()

		dm.walletRepoMock.
			EXPECT().
			CreateAllocations(
				mock.Anything,
				cmd.WalletID,
				mock.Anything,
				mock.Anything,
			).
			RunAndReturn(func(
				ctx context.Context,
				wID uuid.UUID,
				spec wallet.ProviderAllocationSpec,
				updateFunc func(*wallet.Wallet) error,
			) error {
				pa, err := wallet.NewFpAllocation(provider1, 40)
				require.NoError(t, err)

				w, err := wallet.UnmarshalWalletFromDatabase(
					wID,
					"Tai chinh tong",
					0,
					"USD",
					0,
					pa,
				)
				require.NoError(t, err)

				return updateFunc(w)
			})

		err = dm.NewHandler().Handle(context.Background(), cmd)
		require.Error(t, err)
	})

	t.Run("returns error when GetByIDs return nil", func(t *testing.T) {
		provider1, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		provider2, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: provider1.ID(), AllocatedAmount: 50},
				{ID: provider2.ID(), AllocatedAmount: 50},
			},
		}

		dm := NewAllocateFundDM(t)

		dm.fundProviderRepoMock.
			EXPECT().
			GetByIDs(mock.Anything, mock.Anything).
			Return(
				[]*fundprovider.FundProvider{
					provider1,
					nil,
				},
				nil,
			).
			Once()

		err = dm.NewHandler().Handle(context.Background(), cmd)
		require.Error(t, err)
	})

	t.Run("allocate fund provider successfully", func(t *testing.T) {
		provider1, err := fundprovider.NewFundProvider("Techcombank", "BANK", 100, "USD")
		require.NoError(t, err)

		cmd := command.AllocateFundCmd{
			WalletID: uuid.New(),
			AllocationProviders: []command.AllocatedProvider{
				{ID: provider1.ID(), AllocatedAmount: 50},
			},
		}

		dm := NewAllocateFundDM(t)

		// 1. Mock GetByIDs
		dm.fundProviderRepoMock.
			EXPECT().
			GetByIDs(mock.Anything, mock.Anything).
			Return(
				[]*fundprovider.FundProvider{
					provider1,
				},
				nil,
			).
			Once()

		dm.walletRepoMock.
			EXPECT().
			CreateAllocations(
				mock.Anything,
				cmd.WalletID,
				mock.Anything,
				mock.Anything,
			).
			RunAndReturn(func(
				ctx context.Context,
				wID uuid.UUID,
				spec wallet.ProviderAllocationSpec,
				updateFunc func(*wallet.Wallet) error,
			) error {
				w, err := wallet.NewWallet("USD", "Tai chinh tong")
				require.NoError(t, err)

				return updateFunc(w)
			})

		err = dm.NewHandler().Handle(context.Background(), cmd)
		require.NoError(t, err)
	})
}
