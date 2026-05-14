package shared_test

import (
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertValidMoney(t *testing.T, amount decimal.Decimal, currency shared.Currency) shared.Money {
	t.Helper()
	money, err := shared.NewMoney(amount, currency)
	require.NoError(t, err)
	return money
}

func TestNewMoney_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		amount         decimal.Decimal
		currency       shared.Currency
		wantErrDetails []common.ErrorDetails
	}{
		{
			name:     "zero-amount-reject",
			amount:   decimal.Decimal{},
			currency: shared.MustNewCurrency("VND"),
			wantErrDetails: []common.ErrorDetails{
				{
					EntityType: "money",
					ErrorSlug:  "amount-empty",
					Message:    "amount can't be empty",
				},
			},
		},
		{
			name:     "zero-currency-reject",
			amount:   decimal.NewFromInt(100_000),
			currency: shared.Currency{},
			wantErrDetails: []common.ErrorDetails{
				{
					EntityType: "money",
					ErrorSlug:  "currency-empty",
					Message:    "currency can't be empty",
				},
			},
		},
		{
			name:     "zero-amount-and-zero-currency-reject",
			amount:   decimal.Decimal{},
			currency: shared.Currency{},
			wantErrDetails: []common.ErrorDetails{
				{
					EntityType: "money",
					ErrorSlug:  "amount-empty",
					Message:    "amount can't be empty",
				},
				{
					EntityType: "money",
					ErrorSlug:  "currency-empty",
					Message:    "currency can't be empty",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := shared.NewMoney(tt.amount, tt.currency)
			require.Error(t, err)
			testutils.AssertErrorDetails(t, err, tt.wantErrDetails)
		})
	}
}

func TestMoney_Add(t *testing.T) {
	t.Run("returns error when currencies do not match", func(t *testing.T) {
		amount1 := decimal.NewFromInt(100_000)
		amount2 := decimal.NewFromInt(200_000)
		vnd := shared.MustNewCurrency("VND")
		krw := shared.MustNewCurrency("KRW")

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount2, krw)

		_, err := money1.Add(money2)
		require.Error(t, err)

		var expectErr shared.ErrCurrencyMismatch
		assert.ErrorAs(t, err, &expectErr)
	})

	t.Run("adds money successfully when currencies match", func(t *testing.T) {
		amount1 := decimal.NewFromInt(100_000)
		amount2 := decimal.NewFromInt(200_000)
		vnd := shared.MustNewCurrency("VND")

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount2, vnd)

		total, err := money1.Add(money2)
		require.NoError(t, err)

		expectTotal := assertValidMoney(t, amount1.Add(amount2), vnd)

		assert.True(t, total.Equal(expectTotal))
	})
}

func TestMoney_Sub(t *testing.T) {
	t.Run("returns error when currencies do not match", func(t *testing.T) {
		t.Parallel()

		amount1 := decimal.NewFromInt(100_000)
		amount2 := decimal.NewFromInt(200_000)
		vnd := shared.MustNewCurrency("VND")
		krw := shared.MustNewCurrency("KRW")

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount2, krw)

		_, err := money1.Sub(money2)
		require.Error(t, err)

		var expectErr shared.ErrCurrencyMismatch
		assert.ErrorAs(t, err, &expectErr)
	})

	t.Run("subs money successfully when currencies match", func(t *testing.T) {
		t.Parallel()

		amount1 := decimal.NewFromInt(200_000)
		amount2 := decimal.NewFromInt(100_000)
		vnd := shared.MustNewCurrency("VND")

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount2, vnd)

		total, err := money1.Sub(money2)
		require.NoError(t, err)

		expectTotal := assertValidMoney(t, decimal.NewFromInt(100_000), vnd)

		assert.True(t, total.Equal(expectTotal))
	})

	t.Run("subs money successfully when currencies match and right money is greater than left money", func(t *testing.T) {
		t.Parallel()

		leftAmount := decimal.NewFromInt(100_000)
		rightAmount := decimal.NewFromInt(200_000)
		vnd := shared.MustNewCurrency("VND")

		leftMoney := assertValidMoney(t, leftAmount, vnd)

		rightMoney := assertValidMoney(t, rightAmount, vnd)

		total, err := leftMoney.Sub(rightMoney)
		require.NoError(t, err)

		expectTotal := assertValidMoney(t, decimal.NewFromInt(-100_000), vnd)

		assert.True(t, total.Equal(expectTotal))
	})
}

func TestMoney_ZeroValue(t *testing.T) {
	t.Parallel()

	zeroMoney := shared.Money{}

	assert.True(t, zeroMoney.IsZero())
}

func TestMoney_Equal(t *testing.T) {
	amount1 := decimal.NewFromInt(100_000)
	amount2 := decimal.NewFromInt(200_000)
	vnd := shared.MustNewCurrency("VND")
	krw := shared.MustNewCurrency("KRW")

	t.Run("should equal", func(t *testing.T) {
		t.Parallel()

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount1, vnd)

		assert.True(t, money1.Equal(money2))
	})

	t.Run("should not equal when different amount", func(t *testing.T) {
		t.Parallel()

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount2, vnd)

		assert.False(t, money1.Equal(money2))
	})

	t.Run("should not equal when currency different", func(t *testing.T) {
		t.Parallel()

		money1 := assertValidMoney(t, amount1, vnd)

		money2 := assertValidMoney(t, amount1, krw)

		assert.False(t, money1.Equal(money2))
	})
}

func TestMoney_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	moneyJson := `
	{
		"amount": 200000,
		"currency": "VND"	
	}
	`

	var money shared.Money
	err := money.UnmarshalJSON([]byte(moneyJson))
	require.NoError(t, err)

	assert.Equal(t, money.Amount(), decimal.NewFromInt(200_000))
	assert.True(t, money.Currency().Equal(shared.MustNewCurrency("VND")))
}

func TestMoney_MarshalJSON(t *testing.T) {
	t.Parallel()

	amount := decimal.NewFromInt(200_000)
	vnd := shared.MustNewCurrency("VND")

	money := assertValidMoney(t, amount, vnd)

	data, err := money.MarshalJSON()
	require.NoError(t, err)

	assert.JSONEq(t, `{"amount":"200000","currency":"VND"}`, string(data))
}

func TestMoney_InSharedTypes(t *testing.T) {
	t.Parallel()

	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.Money); ok {
			found = true
			break
		}
	}
	require.True(t, found, "Money{} should be in SharedTypes slice")
}
