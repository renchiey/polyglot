# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Monorepo starter: an Expo (React Native) client in `mobile/` and a Go API in `server/`, talking over JSON/HTTP with JWT bearer auth. The Expo app targets iOS, Android, **and web** (Expo Router with `output: "static"`).

## Commands

All `make` targets run from the repo root. `sqlc` and `migrate` must be on `PATH` (`go install` puts them in `$(go env GOPATH)/bin`).

```bash
make db-up         # start Postgres in Docker
make migrate-up    # apply migrations
make sqlc          # regenerate server/internal/db/gen from SQL — REQUIRED after editing queries/ or migrations/
make server        # run API on :8080 (cd server && go run ./cmd/api)
make mobile        # expo start (press w=web, i=iOS, a=android)
make migrate-new name=add_widgets   # scaffold a new migration pair
```

Server-only, from `server/`: `go build ./...`, `go vet ./...`, `go test ./...` (single test: `go test ./internal/handlers -run TestName`).
Mobile-only, from `mobile/`: `npx tsc --noEmit` (typecheck), `npm run lint`.

Both `server/.env` and `mobile/.env` are required at runtime — copy from the `.env.example` files. The server fails fast if `DATABASE_URL` or `JWT_SECRET` is unset.

## Architecture

**Server** (`server/`, module `github.com/renchieyang/language-app/server`) — chi router. `cmd/api/main.go` loads config, opens a pgx pool, builds the router (`internal/server`), and runs with graceful shutdown. Request flow: chi middleware → route → `handlers` → `gen.Queries` (sqlc) → Postgres. `internal/auth` issues/parses HS256 JWTs and provides the bearer-token middleware that injects the user ID into the request context (read via `auth.UserID(ctx)`). Passwords are bcrypt-hashed in `handlers/auth.go`. JSON responses/errors go through `internal/httputil`.

**Database layer is sqlc-generated.** Do not hand-edit `internal/db/gen/` — it is regenerated from `internal/db/queries/*.sql` (the queries) and `internal/db/migrations/*.up.sql` (the schema sqlc reads). `sqlc.yaml` sets two type overrides that the handlers depend on: Postgres `uuid` → `github.com/google/uuid.UUID` and `timestamptz` → `time.Time` (instead of pgtype). After changing any query or migration, run `make sqlc` or the server won't compile against the new shapes.

**Migrations** (golang-migrate) are the source of truth for schema. Files are **6-digit zero-padded** (`000001_init`, `000002_add_words`, …) — match that width (`make migrate-new` does), since mixed widths that resolve to the same integer version make migrate fail with a "duplicate migration" error. Each step is an `.up.sql` / `.down.sql` pair. sqlc reads the same migration files as its schema, so schema lives in exactly one place.

**FSRS scheduling** — the `cards` table stores a `github.com/open-spaced-repetition/go-fsrs/v4` `Card` per `(user, word)`. The generated `gen.Card` is **not** identical to `fsrs.Card`: Postgres has no unsigned ints (uint64 counters become `int64`/`int32`) and models an unreviewed card's `last_review` as `NULL` (`*time.Time`) vs the scheduler's zero `time.Time`. `internal/srs` is the single place that converts both directions (`ToUpsertParams`, `FromRow`) — never scan a `gen.Card` straight into scheduler logic; go through it.

**Mobile** (`mobile/`, Expo Router) — file-based routing under `app/`; `app/_layout.tsx` loads design-system fonts (holding the splash screen until ready) and wraps everything in `SafeAreaProvider` + `AuthProvider`, then defines the Stack. Source lives under `src/` (aliased `@/*` in tsconfig). `src/api/client.ts` is a typed `fetch` wrapper that reads the API base URL from `process.env.EXPO_PUBLIC_API_URL` and throws `ApiError` on non-2xx; `src/api/auth.ts` holds endpoint functions + shared types. `src/auth/AuthContext.tsx` owns auth state, persisting the JWT and restoring the session on mount by calling `/me`. `src/lib/storage.ts` is the **platform split that matters**: SecureStore on native, `localStorage` on web — keep new persisted state going through it so web builds don't break.

**Design system** — the project follows the "Playful Geometric" system (full spec in Claude's project memory). Tokens live in `src/theme/` (`colors`, `typography`, `layout`, `shadows`, `motion`) and are re-exported from `src/theme/index.ts` — import from `@/theme`, never hardcode hex/spacing in screens. Two conventions worth knowing: the signature hard "pop" shadow is produced by `popShadow()`/`cardShadow()`, which emit RN's `boxShadow` string (works on new-arch native + web; offset auto-halves on native per the responsive rule) — its return type is a minimal `{ boxShadow }` so it's usable in both `ViewStyle` and `TextStyle` arrays. Fonts (Outfit headings, Plus Jakarta Sans body) are registered via `useAppFonts()`; the `fontFamily` keys in `typography.ts` must match the names loaded in `fonts.ts`. Reusable primitives go in `src/components/` (`Button`, `Card` exist); build new UI by composing these rather than restyling per screen.

## Gotchas

- The Go module path (`github.com/renchieyang/...`) is a placeholder; if the repo moves, update `go.mod` and all imports.
- Physical devices can't reach `localhost` — set `EXPO_PUBLIC_API_URL` to the host LAN IP in `mobile/.env`.
- The server's `CORS_ORIGINS` must include the web client's dev origin (defaults cover Expo web ports).
