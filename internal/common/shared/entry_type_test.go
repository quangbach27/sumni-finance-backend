package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
)

func TestEntryType_Vars(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "in", shared.EntryTypeIn.String())
	assert.Equal(t, "out", shared.EntryTypeOut.String())
	assert.False(t, shared.EntryTypeIn.IsZero())
	assert.False(t, shared.EntryTypeOut.IsZero())
}

func TestEntryType_ZeroValue(t *testing.T) {
	t.Parallel()

	var e shared.EntryType
	assert.True(t, e.IsZero())
}

func TestEntryType_Reverse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input shared.EntryType
		want  shared.EntryType
	}{
		{
			name:  "debit reverses to credit",
			input: shared.EntryTypeIn,
			want:  shared.EntryTypeOut,
		},
		{
			name:  "credit reverses to debit",
			input: shared.EntryTypeOut,
			want:  shared.EntryTypeIn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want.String(), tt.input.Reverse().String())
		})
	}
}

func TestEntryType_InSharedTypes(t *testing.T) {
	t.Parallel()

	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.EntryType); ok {
			found = true
			break
		}
	}
	require.True(t, found, "EntryType{} should be in SharedTypes slice")
}
