package testutils

import (
	"math/rand/v2"
	"testing"

	"sumni-finance-backend/internal/common/shared"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

func RandomBankInfo(t *testing.T) shared.BankInfo {
	t.Helper()
	bankInfo, err := shared.NewBankInfo(
		gofakeit.Company(),
		gofakeit.Number(100000, 999999),
		gofakeit.LetterN(3),
		gofakeit.LetterN(6),
		gofakeit.URL(),
		gofakeit.URL(),
		gofakeit.Number(0, 1),
		gofakeit.Number(0, 1),
	)
	require.NoError(t, err)
	return bankInfo
}
