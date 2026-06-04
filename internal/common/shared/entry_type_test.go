package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
)

func TestEntryType_Vars(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "DEBIT", shared.EntryTypeDebit.String())
	assert.Equal(t, "CREDIT", shared.EntryTypeCredit.String())
	assert.False(t, shared.EntryTypeDebit.IsZero())
	assert.False(t, shared.EntryTypeCredit.IsZero())
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
			input: shared.EntryTypeDebit,
			want:  shared.EntryTypeCredit,
		},
		{
			name:  "credit reverses to debit",
			input: shared.EntryTypeCredit,
			want:  shared.EntryTypeDebit,
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
	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.EntryType); ok {
			found = true
			break
		}
	}
	require.True(t, found, "EntryType{} should be in SharedTypes slice")
}
