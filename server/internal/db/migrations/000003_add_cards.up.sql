-- An FSRS scheduling card, one per (user, word). Columns mirror
-- github.com/open-spaced-repetition/go-fsrs/v4 Card; uint64 fields become
-- signed Postgres integers (Postgres has no unsigned type).
CREATE TABLE cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_id         UUID NOT NULL REFERENCES words(id) ON DELETE CASCADE,

    due             TIMESTAMPTZ      NOT NULL,
    stability       DOUBLE PRECISION NOT NULL,
    difficulty      DOUBLE PRECISION NOT NULL,
    scheduled_days  BIGINT           NOT NULL DEFAULT 0,
    reps            INTEGER          NOT NULL DEFAULT 0,
    lapses          INTEGER          NOT NULL DEFAULT 0,
    state           SMALLINT         NOT NULL DEFAULT 0,
    last_review     TIMESTAMPTZ,                       -- null until first review
    remaining_steps INTEGER          NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, word_id)
);

-- Hot path: "which cards are due for this user right now".
CREATE INDEX cards_user_due_idx ON cards (user_id, due);
