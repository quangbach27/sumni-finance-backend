BEGIN;

CREATE SCHEMA IF NOT EXISTS identities;

CREATE TABLE identities.sessions (    
    session_id          UUID NOT NULL,
    user_subject        TEXT NOT NULL,
    id_token            TEXT NOT NULL,
    access_token        TEXT NOT NULL,
    refresh_token       TEXT NOT NULL,
    active_organization varchar(255) NOT NULL,
    active_office       varchar(255) NOT NULL,
    is_expired          boolean NOT NULL,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at           TIMESTAMPTZ,
    PRIMARY KEY (session_id)
);

CREATE UNIQUE INDEX idx_unique_active_user_session 
ON identities.sessions (user_subject) 
WHERE is_expired = FALSE;

COMMIT;