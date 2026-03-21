package command_test

import (
	"context"
	"sumni-finance-backend/internal/finance/app/command"
	fp_mock "sumni-finance-backend/internal/finance/domain/fundprovider/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CreateFundProviderDependenciesManager struct {
	fundProviderRepoMock *fp_mock.MockRepository
}

func NewCreateFundProviderDM(t *testing.T) *CreateFundProviderDependenciesManager {
	t.Helper()

	return &CreateFundProviderDependenciesManager{
		fundProviderRepoMock: fp_mock.NewMockRepository(t),
	}
}

func (dm *CreateFundProviderDependenciesManager) NewHandler() command.CreateFundProviderHandler {
	return command.NewCreateFundProviderHandler(dm.fundProviderRepoMock)
}

func TestCreateFundProvider_Handle(t *testing.T) {
	tests := []struct {
		name          string
		cmd           command.CreateFundProviderCmd
		setupMock     func(*CreateFundProviderDependenciesManager)
		hasErr        bool
		errorContains string
	}{
		{
			name: "returns error when name is empty",
			cmd: command.CreateFundProviderCmd{
				Name:         "",
				FpType:       "BANK",
				InitBalance:  100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				// No mock setup needed
			},
			hasErr: true,
		},
		{
			name: "returns error when initBalance is negative",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "BANK",
				InitBalance:  -100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				// No mock setup needed
			},
			hasErr: true,
		},
		{
			name: "returns error when currencyCode is empty",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "BANK",
				InitBalance:  0,
				CurrencyCode: "",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				// No mock setup needed
			},
			hasErr: true,
		},
		{
			name: "returns error when type is empty",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "",
				InitBalance:  100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				// No mock setup needed
			},
			hasErr: true,
		},
		{
			name: "returns error when type is invalid",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "INVALID",
				InitBalance:  100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				// No mock setup needed
			},
			hasErr: true,
		},
		{
			name: "returns error when repository save fails",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "BANK",
				InitBalance:  100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				dm.fundProviderRepoMock.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			hasErr: true,
		},
		{
			name: "creates fund provider successfully when init balance is zero",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "BANK",
				InitBalance:  0,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				dm.fundProviderRepoMock.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			hasErr: false,
		},
		{
			name: "creates fund provider successfully with CASH type",
			cmd: command.CreateFundProviderCmd{
				Name:         "Office Cash",
				FpType:       "CASH",
				InitBalance:  5000,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				dm.fundProviderRepoMock.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			hasErr: false,
		},
		{
			name: "creates fund provider successfully with positive balance",
			cmd: command.CreateFundProviderCmd{
				Name:         "Techcombank7316",
				FpType:       "BANK",
				InitBalance:  100,
				CurrencyCode: "USD",
			},
			setupMock: func(dm *CreateFundProviderDependenciesManager) {
				dm.fundProviderRepoMock.
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
			// Arrange
			dm := NewCreateFundProviderDM(t)
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
