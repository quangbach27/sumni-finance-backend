//go:build integration

package db_test

import (
	"testing"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/identity/adapters/db"
	"sumni-finance-backend/internal/identity/app/models"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertSession(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	session := models.Session{
		ID:                 models.SessionID{UUID: common.NewUUIDv7()},
		UserSubject:        testutils.RandomString(20),
		IDToken:            testutils.RandomString(20),
		AccessToken:        testutils.RandomString(20),
		RefreshToken:       testutils.RandomString(20),
		ActiveOrganization: gofakeit.Company(),
		ActiveOffice:       gofakeit.DomainName(),
		IsExpired:          false,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(5 * time.Minute),
	}

	pool, cleanUp := testutils.NewDB(ctx)
	defer cleanUp()

	sessionRepo := db.NewSessionRepository(pool)

	err := sessionRepo.UpsertSession(ctx, session)
	require.NoError(t, err)

	session2, err := sessionRepo.GetActiveSessionByID(ctx, session.ID)
	require.NoError(t, err)

	diff := cmp.Diff(session, session2, cmpopts.EquateApproxTime(time.Second))
	if diff != "" {
		t.Errorf("sessions do not match (-want +got): \n%s", diff)
	}
}

func TestExpireSession(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	session := models.Session{
		ID:                 models.SessionID{UUID: common.NewUUIDv7()},
		UserSubject:        testutils.RandomString(20),
		IDToken:            testutils.RandomString(20),
		AccessToken:        testutils.RandomString(20),
		RefreshToken:       testutils.RandomString(20),
		ActiveOrganization: gofakeit.Company(),
		ActiveOffice:       gofakeit.DomainName(),
		IsExpired:          false,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(5 * time.Minute),
	}

	pool, cleanUp := testutils.NewDB(ctx)
	defer cleanUp()

	sessionRepo := db.NewSessionRepository(pool)

	err := sessionRepo.UpsertSession(ctx, session)
	require.NoError(t, err)

	err = sessionRepo.ExpireSessionByID(ctx, session.ID)
	require.NoError(t, err)

	expiredSession, err := sessionRepo.GetActiveSessionByID(ctx, session.ID)
	require.NoError(t, err)

	assert.True(t, expiredSession.IsExpired)
}
