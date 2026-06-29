# Handover — Context (AI-native Mandarin learning app)

This orients a fresh session. Read alongside `CLAUDE.md` (repo conventions),
`ROADMAP.md` (phased status), and the auto-memory files under
`~/.claude/projects/.../memory/` (linked from `MEMORY.md`). When in doubt, the
build/tests are the source of truth.

## 1. What this is

Monorepo for "Context", an AI-native Mandarin (HSK 3.0) learning app grounded in
SLA science (comprehensible input i+1, spaced retrieval, active recall,
negotiation of meaning). Three parts:

- **`server/`** — Go API (chi, pgx, sqlc, JWT). This is where almost all work has happened.
- **`web/`** — React + TypeScript + Tailwind v4 client. The primary, fully-featured UI.
- **`mobile/`** — Expo app, basically untouched (auth screen only). Ignore unless asked.

## 2. Status — what's built & working (phases 0–5 + extras)

All verified end-to-end against a local Ollama (`gemma4:26b`) and Postgres.

- **Auth**: JWT register/login/me. (No real auth UI on web yet — see dev shim below.)
- **Vocabulary + SRS (FSRS)**: `words` + `cards`; word CRUD; due queue; review → schedule. Re-adding an existing word **resets its card** (no duplicates).
- **Lexical Auditor pipeline** (`internal/generate`): LLM generate → `lexaudit` HSK audit → correct loop. Provider-agnostic LLM (`internal/llm`: `mock` + `ollama`; anthropic/openai stubbed). `POST /generate`.
- **Mandarin enrichment**: embedded CC-CEDICT (`internal/dict`) → `GET /lookup` (pinyin tone-marks, glosses, char breakdown); `POST /segment` (gse tokens + pinyin); `POST /recall` (generative testing — fresh sentence using a word).
- **Linguistic Elo** (`internal/elo`, `internal/progress`): `linguistic_elo` table (vocab/syntax/listening/speaking). **Only Vocabulary is live** (fed by SRS reviews + translation assessment). Scale anchored so **rating 750 = HSK 1, ~150 Elo ≈ one band**. `GET /progress`. `/generate`+`/recall`+`/journey` derive level from Vocabulary Elo when not specified.
- **Daily Journey** (learner-driven, `internal/journey` + web `JourneyView`): **Read** (full story, pinyin off, tap to mark unknowns) → **Learn** (marked words + a few LLM topic **suggestions**; each looked up & added to vault) → **Recall** (per word: fresh `/recall` sentence, meaning-cued, type-or-reveal → SRS) → **Talk** (stub; voice is Phase 6). `POST /journey/start`, `GET /journey/{id}` (story only; persisted in `journeys` JSONB).
- **Reading Room** (`ReadView`): multi-select **topic chips** → each generate picks a random topic + random style + Elo level; pinyin off; tap word → save to vault; **per-sentence read-aloud** (browser Web Speech API, Chrome) with pronunciation %; **translate box** → `POST /assess/translation` (LLM grades comprehension, nudges Vocab Elo).
- **TTS**: Piper (`internal/tts`, `POST /tts`) — **zero-config auto-discovery** (`make tts-setup`), browser fallback. Web `useSpeak()` everywhere.

## 3. Run & test

```bash
make db-up && make migrate-up          # Postgres (Docker) + migrations
make tts-setup                         # optional: Piper voice into server/voices/ (gitignored)
make server                            # API :8080 (LLM mock by default)
#   real generation: LLM_PROVIDER=ollama LLM_MODEL=gemma4:26b make server
make web-install && make web           # web client :8081 (in the API's default CORS)
```

- **Ollama** runs at `localhost:11434`, model `gemma4:26b` (has `tools`+`thinking`; the ollama adapter sends `think:false`).
- **Dev auth shim**: the web auto-logs-in `dev@context.app` / `context-dev-001` (`web/src/lib/session.tsx`). All test data piles up on this account. There is NO login UI yet by design.
- **Server checks**: `cd server && go build ./... && go vet ./... && go test ./...`. Tested packages: `elo`, `srs`, `generate`, `journey`, `lexaudit`, `dict`, `handlers/audit`. **No DB-backed handler tests** — verify handlers via live curl smokes.
- **Web checks**: `cd web && npm run build` (= `tsc --noEmit && vite build`).
- Ollama integration test: `OLLAMA_MODEL=gemma4:26b go test ./internal/generate -run Ollama -v -count=1 -timeout 10m`.

## 4. Architecture quick map

**Server flow**: chi middleware → `internal/handlers` → `internal/db/gen` (sqlc) → Postgres. Cross-cutting pieces: `lexaudit` (HSK segmentation/audit), `dict` (CC-CEDICT), `llm` (providers), `generate` (pipeline), `srs` (FSRS), `elo`+`progress` (ratings), `journey` (story builder), `tts` (Piper). Router wiring: `internal/server/server.go`.

**DB tables** (migrations `000001`–`000005`): `users`, `words`, `cards` (FSRS), `linguistic_elo`, `journeys` (JSONB content). Schema lives ONLY in migrations; sqlc reads them.

**Endpoints** (all `Authorization: Bearer` except health/auth):
`GET /health` · `POST /auth/register|login` · `GET /me` · `GET /progress` ·
`POST /audit` `/audit/batch` · `POST /generate` · `/words` CRUD · `GET /cards/due` `POST /cards/review` ·
`GET /lookup` · `POST /segment` · `POST /recall` · `POST /assess/translation` · `POST /tts` ·
`POST /journey/start` `GET /journey/{id}`.

**Web** (`web/src/`): `lib/` (api.ts typed client, session.tsx dev-auth, useAsync.ts, tts.ts `useSpeak`); `components/` (ui.tsx primitives, AppShell, PageHeader, LookupSheet); `features/` (Dashboard/Journey/Read/Study/Vault); `App.tsx` (router + connection gate), `index.css` (design tokens).

**Design language**: **block-based, adult/educational** — Space Grotesk (display) + Inter (body) + IBM Plex Mono (tags), flat white bordered blocks, small radii, faint graph-paper background, indigo spine + color-coded accents. NOTE: CSS class names still use the legacy `clay-*` prefix but render **flat** — do NOT reintroduce puffy claymorphism shadows/gradients/Baloo font. Tokens in `web/src/index.css`.

## 5. Conventions & gotchas (will save you time)

- **sqlc**: after editing `internal/db/queries/*.sql` or migrations, run `make sqlc`. The editor LSP shows stale `undefined: gen.X` errors until it reloads the regenerated package — **trust `go build`**, not the inline diagnostics, right after a regen.
- **Migrations** are 6-digit zero-padded `.up.sql`/`.down.sql` pairs; mixed widths break golang-migrate.
- **`go run` orphan**: `make server`/`go run ./cmd/api &` then killing the parent leaves the compiled child holding `:8080`. Use `lsof -ti tcp:8080 | xargs kill -9`, or build to `/tmp/api` and run that, to avoid hitting a stale binary.
- **Pinyin is off by default** in all reading surfaces (learner reads characters first); toggle to show.
- **30s request timeout**: chi `middleware.Timeout(30s)` can clip slow multi-round LLM generations. Fine for single sentences; a per-route/async transport is a known follow-up (needed for voice).
- **Headless screenshots** (how the UI was verified): Chrome `--headless=new` enforces a ~500px min window, so `--window-size=430 --screenshot` clips a wider render — for true mobile use CDP `Emulation.setDeviceMetricsOverride`. Also Chrome's `innerText` returns CSS-**uppercased** pill text, so match case-insensitively when polling page text. (CDP-over-websocket helper snippets are in the session history.)
- **Elo only moves Vocabulary** today; syntax/listening/speaking are seeded at 750 and await grammar/voice signals.
- **Memory system**: durable project facts live in the memory files indexed by `MEMORY.md` — read them; update them when you change those areas.

## 6. Where to go next (priority order)

From `ROADMAP.md` §6 and open follow-ups:

1. **Phase 6 — Voice conversation agent** (fills the Journey "Talk" stub; the headline next feature). Needs STT + an LLM dialogue agent constrained to the story context that negotiates meaning, feeding **Listening + Speaking** Elo. TTS (Piper) and a browser read-aloud check already exist; the conversation loop + transport do not. This forces the **streaming/async transport** decision (the 30s timeout won't fit voice turns).
2. **Phase 7 — Boss Fights**: task-completion simulations gated by Elo thresholds; failures pipe back into SRS.
3. **Cross-cutting, mostly small**:
   - Real cloud LLM adapters behind `llm.New` (Anthropic default, OpenAI) — interface is ready; today only `mock`+`ollama`.
   - Replace the web dev-auth shim with a real auth flow.
   - Persist/cache generated content (avoid re-generating; enable history).
   - Add a `(user_id, term)` uniqueness constraint to back the vault dedup at the DB level (currently handled in the handler).
   - Syntax Elo signal (e.g., grammar/particle checks in the audit).
   - Mobile (`mobile/`) is far behind the web client.
   - TTS caveat: `make tts-setup` installs into the host's `pip`/`python3`; auto-detection of `python3 -m piper` only works when the server runs in that same environment. Bundling a standalone piper binary would remove that.

## 7. How we work here (process notes)

- Build features in vertical slices, verify with `go build/vet/test` + a live curl/CDP smoke against real Ollama, then update `ROADMAP.md` and the relevant memory file.
- Reuse before adding: there are shared helpers (`handlers.userUUID`, `handlers.applyReview`, `resolveAuditor`, `progress.Tracker`, `srs.Scheduler`, web `useSpeak`/`useAsync`/`ui.tsx`). Match existing patterns.
- The user prefers concise, decisive work and honest reporting (say what's verified vs stubbed).
