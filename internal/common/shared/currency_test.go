package shared_test

import (
	"testing"

	"sumni-finance-backend/internal/common/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustNewCurrency(t *testing.T) {
	t.Parallel()

	validCurrencies := []string{"VND", "KRW"}

	for _, code := range validCurrencies {
		t.Run(code, func(t *testing.T) {
			currency := shared.MustNewCurrency(code)
			assert.Equal(t, code, currency.Code())
			assert.False(t, currency.IsZero())
		})
	}
}

func TestMustNewCurrency_InvalidValue(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		shared.MustNewCurrency("INVALID")
	})

	assert.Panics(t, func() {
		shared.MustNewCurrency("vnd")
	})
}

func TestCurrency_ZeroValue(t *testing.T) {
	t.Parallel()

	var c shared.Currency
	assert.True(t, c.IsZero())
	assert.Equal(t, "", c.Code())
}

func TestCurrency_InSharedTypes(t *testing.T) {
	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.Currency); ok {
			found = true
			break
		}
	}
	require.True(t, found, "Currency{} should be in SharedTypes slice")
}
