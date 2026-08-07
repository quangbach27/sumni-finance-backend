package shared_test

import (
	"testing"

	"sumni-finance-backend/internal/common/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTenantContext_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		officeID string
		wantErr  string
	}{
		{
			name:     "empty-tenant-id-reject",
			tenantID: "",
			officeID: "office-1",
			wantErr:  "tenant id can't be empty",
		},
		{
			name:     "empty-office-id-reject",
			tenantID: "tenant-1",
			officeID: "",
			wantErr:  "office id can't be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := shared.NewTenantContext(tt.tenantID, tt.officeID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewTenantContext_Success(t *testing.T) {
	ctx, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)
	assert.Equal(t, "tenant-1", ctx.TenantID())
	assert.Equal(t, "office-1", ctx.OfficeID())
}
