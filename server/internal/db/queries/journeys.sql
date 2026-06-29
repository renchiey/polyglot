-- name: CreateJourney :one
INSERT INTO journeys (user_id, topic, level, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetJourney :one
SELECT * FROM journeys
WHERE id = $1 AND user_id = $2;

-- name: UpdateJourneyContent :one
-- Persist target progress after an answer; flips completed when all are done.
UPDATE journeys
SET content = $3, completed = $4, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;
