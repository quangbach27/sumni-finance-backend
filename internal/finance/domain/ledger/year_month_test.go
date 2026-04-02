package ledger_test

import (
	"testing"

	"sumni-finance-backend/internal/finance/domain/ledger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYearMonth_NewYearMonth(t *testing.T) {
	tests := []struct {
		name   string
		month  int
		year   int
		hasErr bool
	}{
		{
			name:   "returns error when year is zero",
			month:  1,
			year:   0,
			hasErr: true,
		},
		{
			name:   "returns error when year is negative",
			month:  1,
			year:   -1,
			hasErr: true,
		},
		{
			name:   "returns error when month is zero",
			month:  0,
			year:   2024,
			hasErr: true,
		},
		{
			name:   "returns error when month is negative",
			month:  -1,
			year:   2024,
			hasErr: true,
		},
		{
			name:   "returns error when month is greater than 12",
			month:  13,
			year:   2024,
			hasErr: true,
		},
		{
			name:   "creates year month successfully with January",
			month:  1,
			year:   2024,
			hasErr: false,
		},
		{
			name:   "creates year month successfully with December",
			month:  12,
			year:   2024,
			hasErr: false,
		},
		{
			name:   "creates year month successfully with mid-year month",
			month:  6,
			year:   2026,
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ym, err := ledger.NewYearMonth(tt.month, tt.year)

			if tt.hasErr {
				require.Error(t, err)
				assert.True(t, ym.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.year, ym.Year())
				assert.Equal(t, tt.month, ym.Month())
				assert.False(t, ym.IsZero())
			}
		})
	}
}

func TestYearMonth_UnmarshallYearMonthFromDatabase(t *testing.T) {
	testCases := []struct {
		name        string
		ymStr       string
		hasErr      bool
		expectYear  int
		expectMonth int
	}{
		{
			name:   "returns error when string is empty",
			ymStr:  "",
			hasErr: true,
		},
		{
			name:   "returns error when string format is invalid (no comma)",
			ymStr:  "2024",
			hasErr: true,
		},
		{
			name:   "returns error when string has too many parts",
			ymStr:  "2024,3,1",
			hasErr: true,
		},
		{
			name:   "returns error when year is not a number",
			ymStr:  "abc,3",
			hasErr: true,
		},
		{
			name:   "returns error when month is not a number",
			ymStr:  "2024,abc",
			hasErr: true,
		},
		{
			name:   "returns error when year is zero",
			ymStr:  "0,3",
			hasErr: true,
		},
		{
			name:   "returns error when month is invalid (0)",
			ymStr:  "2024,0",
			hasErr: true,
		},
		{
			name:   "returns error when month is invalid (13)",
			ymStr:  "2024,13",
			hasErr: true,
		},
		{
			name:        "unmarshalls successfully with January",
			ymStr:       "2024,1",
			hasErr:      false,
			expectYear:  2024,
			expectMonth: 1,
		},
		{
			name:        "unmarshalls successfully with December",
			ymStr:       "2024,12",
			hasErr:      false,
			expectYear:  2024,
			expectMonth: 12,
		},
		{
			name:        "unmarshalls successfully with whitespace",
			ymStr:       "  2026,6  ",
			hasErr:      false,
			expectYear:  2026,
			expectMonth: 6,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ym, err := ledger.UnmarshalYearMonthFromDatabase(tt.ymStr)

			if tt.hasErr {
				require.Error(t, err)
				assert.True(t, ym.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectYear, ym.Year())
				assert.Equal(t, tt.expectMonth, ym.Month())
				assert.False(t, ym.IsZero())
			}
		})
	}
}

func TestYearMonth_String(t *testing.T) {
	testCases := []struct {
		name         string
		month        int
		year         int
		expectString string
	}{
		{
			name:         "formats January correctly",
			month:        1,
			year:         2024,
			expectString: "2024,1",
		},
		{
			name:         "formats December correctly",
			month:        12,
			year:         2024,
			expectString: "2024,12",
		},
		{
			name:         "formats mid-year month correctly",
			month:        6,
			year:         2026,
			expectString: "2026,6",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ym, err := ledger.NewYearMonth(tt.month, tt.year)
			require.NoError(t, err)

			assert.Equal(t, tt.expectString, ym.String())
		})
	}
}
