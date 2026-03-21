package fundprovider_test

import (
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid BANK type uppercase",
			input:   "BANK",
			want:    "BANK",
			wantErr: false,
		},
		{
			name:    "valid BANK type lowercase",
			input:   "bank",
			want:    "BANK",
			wantErr: false,
		},
		{
			name:    "valid CASH type uppercase",
			input:   "CASH",
			want:    "CASH",
			wantErr: false,
		},
		{
			name:    "valid CASH type lowercase",
			input:   "cash",
			want:    "CASH",
			wantErr: false,
		},
		{
			name:    "valid type with whitespace",
			input:   "  BANK  ",
			want:    "BANK",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fundprovider.NewType(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, fundprovider.ErrInvalidType, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Value())
			}
		})
	}
}
