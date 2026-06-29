-- name: CreateWord :one
INSERT INTO words (user_id, term, translation, definition)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWord :one
SELECT * FROM words
WHERE id = $1 AND user_id = $2;

-- name: GetWordByTerm :one
-- Used to detect re-adding a word already in the vault (so we reset its card
-- instead of creating a duplicate).
SELECT * FROM words
WHERE user_id = $1 AND term = $2
ORDER BY created_at
LIMIT 1;

-- name: ListWordsByUser :many
SELECT * FROM words
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListWordsWithCards :many
-- The vault: each word with its next review time and FSRS state (LEFT JOIN, so
-- next_review is NULL only if a word somehow has no card yet).
SELECT
    w.id, w.user_id, w.term, w.translation, w.definition, w.created_at, w.updated_at,
    c.due AS next_review,
    c.state AS card_state
FROM words w
LEFT JOIN cards c ON c.word_id = w.id AND c.user_id = w.user_id
WHERE w.user_id = $1
ORDER BY w.created_at DESC;

-- name: ListKnownTerms :many
-- Terms whose card is in the given FSRS state (pass Review=2 for "acquired"),
-- used to bias generation and exempt mastered words from the i+1 audit.
SELECT w.term
FROM words w
JOIN cards c ON c.word_id = w.id
WHERE w.user_id = $1 AND c.state = $2
ORDER BY w.created_at DESC;

-- name: UpdateWord :one
UPDATE words
SET term = $3, translation = $4, definition = $5, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteWord :exec
DELETE FROM words
WHERE id = $1 AND user_id = $2;
