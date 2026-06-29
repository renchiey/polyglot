-- A Daily Journey session: the generated story plus its embedded recall targets
-- and interaction handoff, stored as JSON so the shape can evolve without
-- migrations. completed flips once every target has been answered.
CREATE TABLE journeys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic      TEXT NOT NULL DEFAULT '',
    level      INTEGER NOT NULL,
    content    JSONB NOT NULL,
    completed  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX journeys_user_idx ON journeys (user_id, created_at DESC);
