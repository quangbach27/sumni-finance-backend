-- name: UpsertSession :exec
INSERT INTO identities.sessions (
    session_id,
    user_subject,
    id_token,
    access_token,
    refresh_token,
    active_organization,
    active_office,
    is_expired,
    expires_at,
    created_at,
    updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (user_subject) WHERE is_expired = FALSE 
DO UPDATE SET
    session_id = EXCLUDED.session_id,
    id_token = EXCLUDED.id_token,
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    active_organization = EXCLUDED.active_organization,
    active_office = EXCLUDED.active_office,
    expires_at = EXCLUDED.expires_at,
    updated_at = EXCLUDED.updated_at;

-- name: GetActiveSessionByID :one
SELECT * FROM identities.sessions
WHERE session_id = $1 
    AND is_expired = false;

-- name: ExpireSessionByID :exec
UPDATE identities.sessions
SET is_expired = TRUE, updated_at = NOW()
WHERE session_id = $1;