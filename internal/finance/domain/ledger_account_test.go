package domain_test

import (
	"testing"
	"time"

	"sumni-finance-backend/internal/finance/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLedgerPeriodConfig(t *testing.T) {
	t.Run("ValidationErrors", func(t *testing.T) {
		tests := []struct {
			name       string
			dayOfMonth int
			frequency  domain.PeriodFrequency
			wantErr    string
		}{
			{
				name:       "reject-day-below-1",
				dayOfMonth: 0,
				frequency:  domain.PeriodFrequencyMonthly,
				wantErr:    "dayOfMonth must be between 1 and 28",
			},
			{
				name:       "reject-day-above-28",
				dayOfMonth: 29,
				frequency:  domain.PeriodFrequencyMonthly,
				wantErr:    "dayOfMonth must be between 1 and 28",
			},
			{
				name:       "reject-negative-day",
				dayOfMonth: -1,
				frequency:  domain.PeriodFrequencyMonthly,
				wantErr:    "dayOfMonth must be between 1 and 28",
			},
			{
				name:       "reject-zero-frequency",
				dayOfMonth: 15,
				frequency:  domain.PeriodFrequency{},
				wantErr:    "frequency can't be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := domain.NewLedgerPeriodConfig(tt.dayOfMonth, tt.frequency)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		tests := []struct {
			name       string
			dayOfMonth int
			frequency  domain.PeriodFrequency
		}{
			{
				name:       "day-1-monthly",
				dayOfMonth: 1,
				frequency:  domain.PeriodFrequencyMonthly,
			},
			{
				name:       "day-28-quarterly",
				dayOfMonth: 28,
				frequency:  domain.PeriodFrequencyQuarterly,
			},
			{
				name:       "day-15-annually",
				dayOfMonth: 15,
				frequency:  domain.PeriodFrequencyAnnually,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				cfg, err := domain.NewLedgerPeriodConfig(tt.dayOfMonth, tt.frequency)
				require.NoError(t, err)

				assert.Equal(t, tt.dayOfMonth, cfg.DayOfMonth())
				assert.Equal(t, tt.frequency, cfg.Frequency())
			})
		}
	})
}

func TestLedgerPeriodConfig_CalculateCloseDate(t *testing.T) {
	tests := []struct {
		name      string
		from      time.Time
		frequency domain.PeriodFrequency
		wantDate  time.Time
	}{
		{
			name:      "monthly-mid-year",
			from:      time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyMonthly,
			wantDate:  time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "monthly-december-rolls-to-next-year",
			from:      time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyMonthly,
			wantDate:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "quarterly-mid-year",
			from:      time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyQuarterly,
			wantDate:  time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "quarterly-october-rolls-to-next-year",
			from:      time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyQuarterly,
			wantDate:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "annually-mid-year",
			from:      time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyAnnually,
			wantDate:  time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "annually-january-to-next-year",
			from:      time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			frequency: domain.PeriodFrequencyAnnually,
			wantDate:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := domain.NewLedgerPeriodConfig(15, tt.frequency)
			require.NoError(t, err)

			closeDate := cfg.CalculateCloseDate(tt.from)
			assert.Equal(t, tt.wantDate, closeDate)
		})
	}
}
