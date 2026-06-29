# Polyglot — Web Client

A responsive, gamified **claymorphism** web UI for the Polyglot language-learning
API. Built with Vite + React + TypeScript + Tailwind v4.

## Run it

```bash
# 1. Backend (from repo root)
make db-up && make migrate-up
make server                       # API on :8080
# optional: real generations via Ollama
# LLM_PROVIDER=ollama LLM_MODEL=gemma4:26b make server

# 2. Web client
make web-install                  # first time
make web                          # http://localhost:8081
```

The dev server runs on **:8081**, which the API's default `CORS_ORIGINS` already
allows — no backend change needed. Point at a different API with the connection
pill in the header, or set `VITE_API_URL` in `web/.env`.

## Auth

There is **no login screen yet** (by design). On load the app silently signs in
a fixed dev account (`dev@context.app`) — logging in, or registering it once — so
the authenticated endpoints are testable. Swap this for the real auth flow later;
it lives entirely in `src/lib/session.tsx`.

## What you can test

| View | Endpoints exercised |
|------|---------------------|
| **Home** | `/words`, `/cards/due` (live counts) |
| **Read** | `/generate` → `/segment`; tap a word → `/lookup`; "Add to Vault" → `/words`; anchor game → `/anchor` |
| **Study** | `/cards/due`, `/cards/review` (FSRS ratings); Recall quiz → `/recall` |
| **Vault** | `/words` CRUD (+ tap to `/lookup`) |

## Structure

```
src/
  lib/        api.ts (typed client) · session.tsx (dev auth) · useAsync.ts
  components/ ui.tsx (clay primitives) · AppShell · PageHeader
              InteractiveText (poke-able hanzi) · LookupSheet (+ anchor game)
  features/   DashboardView · ReadView · StudyView · VaultView
  index.css   claymorphism design tokens + component classes
```

Design tokens and clay shadows live in `src/index.css`; add new features as a
view in `src/features/` plus a route in `App.tsx` and a nav entry in `AppShell`.
The roadmap surfaces (Elo, Boss Fights, Voice, Daily Journey) appear as
"coming soon" tiles on the dashboard.
