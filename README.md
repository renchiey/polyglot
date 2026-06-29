# Polyglot

An **AI-native Mandarin (HSK 3.0) learning platform**. Instead of static flashcard
decks, Polyglot generates a shifting, individualized stream of text (and soon audio)
that stays just past your current level — driven by LLMs and deterministic NLP, and
grounded in Second Language Acquisition science.

It's a monorepo: a **React + TypeScript web client**, a **Go API**, and an **Expo
(React Native) app** that share one JSON/HTTP + JWT backend.

```
language-app/
├── web/         Vite + React + TypeScript + Tailwind v4 — the primary, full-featured client
├── server/      Go API — chi router, pgx, sqlc, JWT auth, Postgres
├── mobile/      Expo Router app (iOS / Android / Web) — auth screen only so far
├── docker-compose.yml   Postgres for local dev
└── Makefile     common dev commands
```

## Why it works this way (pedagogical pillars)

Every feature serves one of these SLA principles:

- **Comprehensible Input (i+1)** — content is ~90–95% known + ~5% new, at the next tier.
- **Spaced Retrieval** — review a word right as it's about to be forgotten (FSRS).
- **Active Recall** — force production from memory, not multiple-choice recognition.
- **Negotiation of Meaning** — (planned) a voice agent that feigns misunderstanding so the learner self-corrects.

## Current features

Everything below is built and verified end-to-end (against local Ollama + Postgres).

### Learning engine (Go server)

- **Lexical Auditor pipeline** (`internal/generate`) — the novel core. An LLM generates
  text, `internal/lexaudit` deterministically checks it stays within the learner's HSK
  level (gse segmentation + HSK 3.0 lexicon, plus a known-words allowlist for i+1), and a
  correction pass rewrites violations — looping until it passes or a round budget runs out.
- **Provider-agnostic LLM** (`internal/llm`) — `mock` and `ollama` adapters today;
  Anthropic/OpenAI stubbed behind the same interface.
- **Vocabulary + SRS (FSRS)** — `words` + `cards`; word CRUD, a due queue, and reviews
  that reschedule via [`go-fsrs`](https://github.com/open-spaced-repetition/go-fsrs).
  Re-adding a word resets its card (no duplicates).
- **Mandarin enrichment** — embedded CC-CEDICT dictionary (`internal/dict`): `lookup`
  gives pinyin (tone marks), glosses, and a per-character breakdown; `segment` tokenizes a
  passage with per-word pinyin; `recall` is generative testing — a fresh sentence built
  from a due word + only known vocab, returned as a fill-in-the-blank cloze.
- **Linguistic Elo** (`internal/elo`, `internal/progress`) — a fluid rating across four
  vectors (Vocabulary / Syntax / Listening / Speaking), anchored so **750 = HSK 1**
  (~150 Elo ≈ one band). **Vocabulary is live** (fed by SRS reviews + translation
  assessment); generation derives its target level from it when none is given. The other
  three vectors await their signals (grammar checks, voice).
- **Daily Journey** (`internal/journey`) — a learner-driven narrative loop persisted as
  JSONB: an i+1 story (segmented, per-token pinyin) the learner reads, marks unknowns,
  learns, and recalls.
- **Neural TTS** (`internal/tts`) — Mandarin speech via
  [Piper](https://github.com/OHF-Voice/piper1-gpl), zero-config (auto-discovers the voice),
  with playback **speed** (pitch-preserving via Piper length-scale) and **volume**
  configurable from the client. Falls back to the browser's built-in voice when Piper isn't set up.

### Web client (`web/`)

A responsive, block-based UI (Space Grotesk + Inter + IBM Plex Mono) covering every
implemented feature:

- **Home** — live vocabulary / due-card counts and a Linguistic Elo snapshot.
- **Journey** — 4-phase flow: **Read** (story, pinyin off by default, tap to mark
  unknowns) → **Learn** (looks each word up, adds to the vault) → **Recall** (fresh cloze
  sentence → SRS + Elo) → **Talk** (stub; voice is planned).
- **Read (Reading Room)** — pick topic chips → generate an i+1 passage; tap any hanzi for
  a lookup + breakdown and save to the vault; per-sentence read-aloud; a translate box
  graded by the LLM (nudges Vocabulary Elo).
- **Study** — the FSRS due queue with Again/Hard/Good/Easy ratings, plus recall quizzes.
- **Vault** — vocabulary CRUD with tap-to-lookup.
- **Settings** — voice playback speed & volume (saved per device, with a live preview).

> The web client has **no login screen yet** — it auto-signs-in a dev account on load
> (`web/src/lib/session.tsx`). Replacing this with a real auth flow is a planned task.

## Planned features

See [`ROADMAP.md`](ROADMAP.md) for the full phased plan. Headline items:

- **Voice / Conversation (Phase 6)** — the Journey's "Talk" stub. STT/ASR for speaking,
  tonal & pronunciation scoring (which finally feeds the **Listening** and **Speaking**
  Elo vectors), and a voice agent constrained to the story's context that negotiates
  meaning. Forces a streaming/async transport decision (the current 30s request timeout
  won't fit voice turns).
- **Boss Fights (Phase 7)** — real-world task simulations gated by Elo thresholds;
  failures pipe the words/patterns that broke communication back into the Daily Journey.
- **Real cloud LLM adapters** — Anthropic (default) + OpenAI behind the existing
  `llm.New` interface (today only `mock` + `ollama`).
- **Real web auth** — replace the dev-account bootstrap with a proper login flow.
- **Mobile surfaces** — bring the Expo app up to the web client's feature set on the
  existing `@/theme` design system.
- **Cross-cutting** — cache/persist generated content, generation observability, LLM rate
  limiting & cost guardrails, a `(user_id, term)` uniqueness constraint, and CI.

## Tech stack

| Layer | Tech |
|-------|------|
| Web | Vite, React, TypeScript, Tailwind v4, React Router |
| Mobile | Expo Router (React Native), TypeScript |
| API | Go, chi, pgx, sqlc, golang-migrate, JWT (HS256), bcrypt |
| Data | PostgreSQL |
| ML / NLP | go-fsrs (SRS), go-ego/gse + HSK 3.0 (lexical audit), CC-CEDICT (dictionary), Piper (TTS), pluggable LLMs (Ollama today) |

## Prerequisites

- Node 20+ and npm
- Go 1.23+
- Docker + Docker Compose
- [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [`migrate`](https://github.com/golang-migrate/migrate) — `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

(`go install` puts both on `$(go env GOPATH)/bin` — make sure that's on your `PATH`.)

## Quick start

```bash
# 1. Copy env templates
cp server/.env.example server/.env
cp mobile/.env.example mobile/.env

# 2. Start Postgres and apply migrations
make db-up
make migrate-up

# 3. Generate typed DB code from SQL (required before building the server)
make sqlc

# 4. Run the API on http://localhost:8080  (LLM mock by default)
make server
#   real generation against local Ollama:
#   LLM_PROVIDER=ollama LLM_MODEL=gemma4:26b make server

# 5. In another terminal, run the web client on http://localhost:8081
make web-install   # first time only
make web
#   — or the Expo app (press w=web, i=iOS, a=android) —
make mobile
```

### Text-to-speech (optional)

```bash
make tts-setup   # pip-install Piper + download the Mandarin voice into server/voices/
```

Zero-config: the server auto-discovers the voice; no `.env` changes needed. Without it,
the web client uses the browser's built-in Mandarin voice. Overrides (`VOICES_DIR`,
`PIPER_VOICE`, `PIPER_BIN`) are documented in `server/.env.example`.


