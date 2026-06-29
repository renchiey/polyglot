# language-app

Monorepo starter: **React Native (Expo Router, web-enabled)** mobile/web client + **Go (chi + pgx + sqlc)** API with JWT auth and Postgres.

```
language-app/
├── mobile/        Expo Router app (iOS / Android / Web), TypeScript
├── server/        Go API: chi router, pgx, sqlc, JWT auth
├── docker-compose.yml   Postgres (+ optional server) for local dev
└── Makefile       common dev commands
```

## Prerequisites

- Node 20+ and npm (or pnpm/yarn)
- Go 1.23+
- Docker + Docker Compose
- [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [`migrate`](https://github.com/golang-migrate/migrate) CLI — `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

## Quick start

```bash
# 1. Start Postgres
make db-up

# 2. Apply migrations
make migrate-up

# 3. Generate typed DB code from SQL (required before building the server)
make sqlc

# 4. Run the API (http://localhost:8080)
make server

# 5. In another terminal, run a client
make mobile        # Expo: press w for web, i for iOS, a for Android
# — or the dedicated web app (claymorphism UI) —
make web-install   # first time only
make web           # http://localhost:8081, talks to the API on :8080
```

The `web/` client is a responsive React app covering every implemented feature
(Read, Study, Vault) for testing the API end to end. It has no login screen yet
and signs in a dev account automatically — see `web/README.md`.

Copy the env templates before running:

```bash
cp server/.env.example server/.env
cp mobile/.env.example mobile/.env
```

### Text-to-speech (optional)

Neural TTS via [Piper](https://github.com/OHF-Voice/piper1-gpl) is zero-config —
run the setup once and the server auto-discovers the voice; no `.env` changes:

```bash
make tts-setup        # pip install piper-tts + download the Mandarin voice into server/voices/
```

Without it, the web client falls back to the browser's built-in Mandarin voice.
Overrides (`VOICES_DIR`, `PIPER_VOICE`, `PIPER_BIN`) are documented in `server/.env.example`.

## API

| Method | Path             | Auth | Description              |
|--------|------------------|------|--------------------------|
| GET    | `/health`        | no   | Liveness/readiness probe |
| POST   | `/auth/register` | no   | Create user, returns JWT |
| POST   | `/auth/login`    | no   | Login, returns JWT       |
| GET    | `/me`            | yes  | Current user profile     |

Send authenticated requests with `Authorization: Bearer <token>`.

## Notes

- `mobile/src/api/` holds a typed fetch client and auth helpers. Token is stored via `expo-secure-store` on native and `localStorage` on web.
- DB access uses `sqlc`-generated code (`server/internal/db/gen`). Edit SQL in `server/internal/db/queries/`, then re-run `make sqlc`.
