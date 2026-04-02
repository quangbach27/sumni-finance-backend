package command_test

import (
	"context"
	"sumni-finance-backend/internal/finance/app/command"
	wallet_mocks "sumni-finance-backend/internal/finance/domain/wallet/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CreateWalletDependenciesManager struct {
	walletRepoMock *wallet_mocks.MockRepository
}

func NewCreateWalletDependenciesManager(t *testing.T) *CreateWalletDependenciesManager {
	t.Helper()

	return &CreateWalletDependenciesManager{
		walletRepoMock: wallet_mocks.NewMockRepository(t),
	}
}

func (dm *CreateWalletDependenciesManager) NewHandler() command.CreateWalletHandler {
	return command.NewCreateWalletHandler(dm.walletRepoMock)
}

func TestCreateWallet_Handle(t *testing.T) {
	tests := []struct {
		name          string
		cmd           command.CreateWalletCmd
		setupMock     func(*CreateWalletDependenciesManager)
		hasErr        bool
		errorContains string
	}{
		{
			name: "returns error when name is empty",
			cmd: command.CreateWalletCmd{
				Name:         "",
				CurrencyCode: "VND",
			},
			setupMock: func(dm *CreateWalletDependenciesManager) {
				// No mock setup needed - validation fails before repo call
			},
			hasErr: true,
		},
		{
			name: "returns error when currency code is empty",
			cmd: command.CreateWalletCmd{
				Name:         "My Wallet",
				CurrencyCode: "",
			},
			setupMock: func(dm *CreateWalletDependenciesManager) {
				// No mock setup needed - validation fails before repo call
			},
			hasErr: true,
		},
		{
			name: "returns error when currency code is invalid",
			cmd: command.CreateWalletCmd{
				Name:         "My Wallet",
				CurrencyCode: "INVALID",
			},
			setupMock: func(dm *CreateWalletDependenciesManager) {
				// No mock setup needed - validation fails before repo call
			},
			hasErr: true,
		},
		{
			name: "returns error when repository save fails",
			cmd: command.CreateWalletCmd{
				Name:         "My Wallet",
				CurrencyCode: "VND",
			},
			setupMock: func(dm *CreateWalletDependenciesManager) {
				dm.walletRepoMock.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			hasErr: true,
		},
		{
			name: "creates wallet successfully",
			cmd: command.CreateWalletCmd{
				Name:         "My Wallet",
				CurrencyCode: "VND",
			},
			setupMock: func(dm *CreateWalletDependenciesManager) {
				dm.walletRepoMock.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			dm := NewCreateWalletDependenciesManager(t)
			tt.setupMock(dm)
			handler := dm.NewHandler()

			// Act
			err := handler.Handle(context.Background(), tt.cmd)

			// Assert
			if tt.hasErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
