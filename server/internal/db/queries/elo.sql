-- name: EnsureElo :one
-- Get-or-create the user's rating row, returning current values (defaults on
-- first call). The no-op UPDATE makes RETURNING fire even on conflict.
INSERT INTO linguistic_elo (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING *;

-- name: UpdateElo :one
UPDATE linguistic_elo
SET vocabulary = $2, syntax = $3, listening = $4, speaking = $5, updated_at = now()
WHERE user_id = $1
RETURNING *;
