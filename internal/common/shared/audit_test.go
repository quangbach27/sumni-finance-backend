package shared_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
)

func TestAudit_InSharedTypes(t *testing.T) {
	t.Parallel()

	found := false
	for _, st := range shared.SharedTypes {
		if _, ok := st.(shared.Audit); ok {
			found = true
			break
		}
	}
	require.True(t, found, "Audit{} should be in SharedTypes slice")
}

func TestNewAudit_WithActor(t *testing.T) {
	t.Parallel()

	before := time.Now()
	a := shared.NewAudit("user-1")
	after := time.Now()

	assert.Equal(t, "user-1", a.CreatedBy())
	assert.True(t, !a.CreatedAt().Before(before) && !a.CreatedAt().After(after))
	assert.Nil(t, a.UpdatedAt())
	assert.Nil(t, a.UpdatedBy())
}

func TestNewAudit_EmptyActorFallsBackToSystem(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("")

	assert.Equal(t, "system", a.CreatedBy())
}

func TestAudit_SetAuditCreate_WithActor(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")

	before := time.Now()
	a.CaptureCreate("user-2")
	after := time.Now()

	assert.Equal(t, "user-2", a.CreatedBy())
	assert.True(t, !a.CreatedAt().Before(before) && !a.CreatedAt().After(after))
}

func TestAudit_SetAuditCreate_EmptyActorFallsBackToSystem(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")
	a.CaptureCreate("")

	assert.Equal(t, "system", a.CreatedBy())
}

func TestAudit_SetAuditCreate_OverwritesPreviousCreatedBy(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")
	a.CaptureCreate("user-2")

	assert.Equal(t, "user-2", a.CreatedBy())
	assert.Nil(t, a.UpdatedAt())
	assert.Nil(t, a.UpdatedBy())
}

func TestAudit_MarkUpdated(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")

	before := time.Now()
	a.CaptureUpdate("user-2")
	after := time.Now()

	require.NotNil(t, a.UpdatedAt())
	require.NotNil(t, a.UpdatedBy())
	assert.Equal(t, "user-2", *a.UpdatedBy())
	assert.True(t, !a.UpdatedAt().Before(before) && !a.UpdatedAt().After(after))
}

func TestAudit_MarkUpdated_OverwritesPreviousUpdate(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")
	a.CaptureUpdate("user-2")
	a.CaptureUpdate("user-3")

	assert.Equal(t, "user-3", *a.UpdatedBy())
}

func TestAudit_CreatedAtNotChangedAfterMarkUpdated(t *testing.T) {
	t.Parallel()

	a := shared.NewAudit("user-1")
	createdAt := a.CreatedAt()

	a.CaptureUpdate("user-2")

	assert.Equal(t, createdAt, a.CreatedAt())
	assert.Equal(t, "user-1", a.CreatedBy())
}
