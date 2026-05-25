package testutils

import (
	"testing"

	"sumni-finance-backend/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertErrorDetails(t *testing.T, err error, expectErrDetails []common.ErrorDetails) {
	t.Helper()

	var commonErr common.Error
	require.ErrorAs(t, err, &commonErr)

	require.Len(t, expectErrDetails, len(commonErr.Details))

	actualErrDetails := make(map[string]common.ErrorDetails, len(commonErr.Details))
	for _, details := range commonErr.Details {
		actualErrDetails[details.ErrorSlug] = details
	}

	for _, expect := range expectErrDetails {
		actual := actualErrDetails[expect.ErrorSlug]

		if expect.EntityID != "" {
			assert.Equal(t, expect.EntityID, actual.EntityID)
		}

		if expect.EntityType != "" {
			assert.Equal(t, expect.EntityType, actual.EntityType)
		}

		if expect.ErrorSlug != "" {
			assert.Equal(t, expect.ErrorSlug, actual.ErrorSlug)
		}

		if expect.Message != "" {
			assert.Equal(t, expect.Message, actual.Message)
		}
	}
}
