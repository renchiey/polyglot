-- One rating row per user across the four PRD proficiency vectors. Stored as
-- doubles so Elo updates stay smooth; the API rounds for display.
-- Default 750 anchors a brand-new learner at HSK 1 on the elo package's scale
-- (DifficultyForLevel(1)); each ~150 points is roughly one HSK band.
CREATE TABLE linguistic_elo (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    vocabulary DOUBLE PRECISION NOT NULL DEFAULT 750,
    syntax     DOUBLE PRECISION NOT NULL DEFAULT 750,
    listening  DOUBLE PRECISION NOT NULL DEFAULT 750,
    speaking   DOUBLE PRECISION NOT NULL DEFAULT 750,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
