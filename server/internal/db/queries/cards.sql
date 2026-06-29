-- name: UpsertCard :one
-- Create or replace the schedule for a (user, word). Use after every review.
INSERT INTO cards (
    user_id, word_id, due, stability, difficulty,
    scheduled_days, reps, lapses, state, last_review, remaining_steps
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (user_id, word_id) DO UPDATE SET
    due             = EXCLUDED.due,
    stability       = EXCLUDED.stability,
    difficulty      = EXCLUDED.difficulty,
    scheduled_days  = EXCLUDED.scheduled_days,
    reps            = EXCLUDED.reps,
    lapses          = EXCLUDED.lapses,
    state           = EXCLUDED.state,
    last_review     = EXCLUDED.last_review,
    remaining_steps = EXCLUDED.remaining_steps,
    updated_at      = now()
RETURNING *;

-- name: GetCardByWord :one
SELECT * FROM cards
WHERE user_id = $1 AND word_id = $2;

-- name: ListDueCards :many
-- Cards due at or before `due`, soonest first. Pass time.Now() and a page size.
SELECT * FROM cards
WHERE user_id = $1 AND due <= $2
ORDER BY due ASC
LIMIT $3;

-- name: ListDueCardsWithWord :many
-- Due cards joined with their word, soonest first. Drives the study queue.
SELECT
    c.id, c.user_id, c.word_id, c.due, c.stability, c.difficulty,
    c.scheduled_days, c.reps, c.lapses, c.state, c.last_review,
    c.remaining_steps, c.created_at, c.updated_at,
    w.term, w.translation
FROM cards c
JOIN words w ON w.id = c.word_id
WHERE c.user_id = $1 AND c.due <= $2
ORDER BY c.due ASC
LIMIT $3;

-- name: DeleteCard :exec
DELETE FROM cards
WHERE id = $1 AND user_id = $2;
