-- name: OfficeByUUID :one
SELECT *
FROM finances.offices
WHERE office_uuid = $1;

-- name: UpsertOffice :exec
INSERT INTO finances.offices (office_uuid, name)
VALUES 
    ($1, $2)
ON CONFLICT (office_uuid) DO
UPDATE SET
    name = EXCLUDED.name;