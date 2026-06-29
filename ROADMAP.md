# Context — Project Outline & Roadmap

AI-native Mandarin (HSK 3.0) acquisition platform. Replaces static flashcard apps with a
shifting, individualized text/audio ecosystem driven by LLMs + deterministic NLP, grounded
in Second Language Acquisition science.

## 1. Pedagogical pillars (the "why" every feature serves)

- **Comprehensible Input (i+1)** — content is 90–95% known + ~5% new at the next tier.
- **Spaced Retrieval** — review a word right as it's about to be forgotten (FSRS).
- **Active Recall** — force production from memory, not multiple-choice recognition.
- **Negotiation of Meaning** — conversational AI that feigns misunderstanding so the learner self-corrects.

## 2. Product surfaces

- **Daily Journey** — one 15-min narrative loop: Input (AI story) → Embedded Retrieval
  (recall targets injected mid-text) → Interaction (voice agent plays a story character).
- **Practice Range** (open sandbox) — Reading Room (infinite i+1 feed), Chat Room (voice
  personas), The Vault (vocabulary mastery matrix).
- **Linguistic Elo** — fluid MMR across 4 vectors (Vocabulary / Syntax / Listening /
  Speaking); "Boss Fights" = real-world task simulations unlock at thresholds.

## 3. Architecture (current)

Monorepo: Expo client (`mobile/`) ↔ Go API (`server/`) over JSON/HTTP + JWT.

**Backend request flow:** chi middleware → handler → `gen.Queries` (sqlc) → Postgres.
The novel core is the **Lexical Auditor pipeline** (`internal/generate`): an LLM generates
text, `internal/lexaudit` deterministically checks it stays within the learner's HSK level
(+ known-words allowlist for i+1), and a correction pass rewrites violations — looping until
it passes or a round budget runs out. LLM access is provider-agnostic (`internal/llm`).

---

## 4. Status snapshot

| Area | State |
|------|-------|
| Auth (JWT register/login/me) | ✅ Done |
| DB schema: users, words, cards (FSRS) + sqlc | ✅ Done |
| FSRS mapper (`internal/srs`) | ✅ Done (no review endpoints yet) |
| Lexical audit (gse + HSK 3.0), `/audit` | ✅ Done |
| LLM client + Mock + **Ollama** adapter | ✅ Done |
| Generation pipeline + `POST /generate` | ✅ Done |
| Known-words allowlist (i+1) | ✅ Done |
| Word CRUD endpoints | ✅ Done |
| SRS review endpoints | ✅ Done |
| Known/learning split feeds allowlist | ✅ Done |
| Dictionary lookup (pinyin / gloss / sub-char) | ✅ Done (`GET /lookup`) |
| Segmentation + passage pinyin (`POST /segment`) | ✅ Done |
| Generative testing + cloze (`POST /recall`) | ✅ Done |
| Anchor mini-game payload (`GET /anchor`) | ✅ Done |
| Linguistic Elo (Vocabulary vector live) | ✅ Done (`GET /progress`) |
| Daily Journey orchestration | ✅ Done (`/journey/*`, web Journey view) |
| TTS (Piper) + reading/speaking assessment | ✅ Done (`/tts`, `/assess/translation`, browser read-aloud) |
| Voice conversation agent (Interaction phase) | ⬜ Not started |
| Boss fights | ⬜ Not started |
| Web client (claymorphism, all current features) | ✅ Done (`web/`, no auth UI yet) |
| Mobile UI beyond auth | ⬜ Not started |
| Real cloud LLM adapters (Anthropic/OpenAI) | ⬜ Stubbed in factory |

---

## 5. Checklist (phased)

Markers: ✅ done · �doing/next · ⬜ todo

### Phase 0 — Foundation ✅
- [x] Go API skeleton (chi, pgx, sqlc, graceful shutdown)
- [x] JWT auth + bearer middleware (`auth.UserID`)
- [x] Migrations + `users` / `words` / `cards` tables
- [x] FSRS card mapper (`internal/srs`)
- [x] Docker Postgres + Makefile + `.env` flow

### Phase 1 — Lexical Auditor pipeline ✅ (current)
- [x] `internal/lexaudit` — gse segmentation + HSK 3.0 lexicon, `/audit` + `/audit/batch`
- [x] `internal/llm` — provider-agnostic `Client`, `MockClient`, `OllamaClient`, factory
- [x] `internal/generate` — generate→audit→correct loop + §1/§3 prompts
- [x] `POST /generate` (auth, biased by user's vault)
- [x] Known-words allowlist exempts acquired-but-advanced vocab (i+1)
- [x] Verified end-to-end vs local `gemma4:26b`

### Phase 2 — Vocabulary & Study Loop ✅
- [x] Word CRUD handlers: `POST/GET/PUT/DELETE /words` + `GET /words/{id}`
- [x] `GET /cards/due?limit=` — due cards joined with their word (`ListDueCardsWithWord`)
- [x] `POST /cards/review` — `{word_id, rating 1-4}` via `srs.Scheduler` (FSRS `Next`) + upsert
- [x] Auto-create a card when a word is added; review lazily creates one if missing
- [x] Known = card in FSRS Review state (`ListKnownTerms`); feeds `/generate` bias + i+1 allowlist
- [x] FSRS scheduler unit tests; full loop verified live (create→due→review→reschedule)

### Phase 3 — Mandarin enrichment (Assisted Noticing + Generative Testing) ✅
- [x] Dictionary data source: CC-CEDICT (CC BY-SA 4.0) filtered to HSK + single chars,
      numbered pinyin → tone marks, gzipped embed (`internal/dict`, `go generate`)
- [x] `GET /lookup?word=…&language=zh` — pinyin + definitions + per-character breakdown
      (most-senses heuristic picks the everyday reading over archaic ones)
- [x] `POST /segment` — tokens with per-word pinyin (maps a tapped char to its compound;
      also gives whole-passage pinyin). Exposes lexaudit's gse segmenter + `IsCJK`
- [x] `POST /recall` — generative testing: fresh sentence using a due word + known vocab
      only (`GenRequest.MustInclude`, exempt from the audit), returns a fill-in-blank cloze
- [x] `GET /anchor?word=…` — unscramble payload (shuffled chars) + pinyin + definitions

### Phase 4 — Linguistic Elo ✅
- [x] Migration `000004_add_linguistic_elo` (4 vectors per user; default 750 = HSK 1)
- [x] `internal/elo` — pure rating math (Expected/Update, HSK level ↔ difficulty, grade→score) + tests
- [x] `internal/progress.Tracker` — applies updates and derives content level; the one place vectors are nudged
- [x] SRS review → **Vocabulary** Elo (difficulty from the word's HSK level; harder items reward more)
- [x] `GET /progress` — four ratings (rounded) + `recommended_level`
- [x] `/generate` (when `target_level` omitted) and `/recall` derive their level from Vocabulary Elo
- [ ] Syntax / Listening / Speaking vectors await their signals (grammar checks, voice — Phase 6)

### Phase 5 — Daily Journey orchestration ✅ (learner-driven, revised 2026-06-29)
- [x] `journeys` table (JSONB) — migration `000005`; `internal/journey.Builder` generates an i+1 story (segmented, per-token pinyin)
- [x] `POST /journey/start`, `GET /journey/{id}` — story only (target/answer machinery removed)
- [x] Web 4-phase flow: **Read** (pinyin off by default, tap to mark unknowns) → **Learn** (`/lookup` teaches + `/words` adds to vault) → **Recall** (`/recall` fresh sentence, meaning-cued → `/cards/review` SRS+Elo) → **Talk** (stub)
- [x] Reading pinyin off-by-default app-wide; dashboard `/progress` Elo card; `MustInclude` generalized to []string

### Phase 6 — Voice / Conversation (Negotiation of Meaning)
- [ ] TTS for Input phase (audio clips)
- [ ] STT / ASR for Speaking
- [ ] Tonal + pronunciation scoring (drives Listening/Speaking Elo)
- [ ] Voice agent endpoint: persona constrained to story context, feigns misunderstanding
- [ ] Streaming/latency strategy (the 30s chi timeout won't fit voice turns)

### Phase 7 — Boss Fights
- [ ] Milestone definitions + unlock thresholds (Elo-gated)
- [ ] Constrained-persona task simulation (binary success on real-world objective)
- [ ] Failure logging → words/patterns that broke communication → back into Daily Journey

### Web client ✅ (claymorphism, `web/` — Vite + React + Tailwind v4)
- [x] Responsive gamified claymorphism UI; dev-auth shim (no login screen yet) in `src/lib/session.tsx`
- [x] Home dashboard (live `/words` + `/cards/due` counts, roadmap "coming soon" tiles)
- [x] Read: `/generate` → `/segment`, poke-able hanzi → `/lookup` + breakdown, Add to Vault, `/anchor` unscramble game
- [x] Study: `/cards/due` + `/cards/review` (FSRS ratings) and `/recall` quiz with cloze
- [x] Vault: `/words` CRUD + tap-to-lookup
- [ ] Real auth flow (replace the dev-account bootstrap)

### Phase 8 — Mobile surfaces
- [ ] Wire API client for words/cards/generate (extend `mobile/src/api/`)
- [ ] The Vault (vocabulary mastery matrix) screen
- [ ] Reading Room (infinite i+1 feed) with tap-to-notice compound highlighting
- [ ] Study screen (SRS reviews)
- [ ] Daily Journey flow (text → embedded recall → voice)
- [ ] Voice chat UI
- [ ] Build on existing `@/theme` design system + `Button`/`Card` primitives

### Cross-cutting / infra
- [ ] Real cloud LLM adapters (Anthropic Claude default, OpenAI) behind `llm.New` — interface already shaped
- [ ] Per-route timeout or async generation (current global 30s middleware clips slow models)
- [ ] Cache / persist generated content (avoid regenerating, enable review)
- [ ] Generation observability: log rounds, pass-rate, out-of-bounds frequency
- [ ] Rate limiting + cost guardrails on LLM calls
- [ ] CI: build, vet, test; integration tests gated by env (e.g. `OLLAMA_MODEL`)
- [ ] Seed/import full HSK vocabulary as starter `words` (so the allowlist is meaningful day one)

---

## 6. Immediate next step

**Phase 6, Voice / Conversation** — the Daily Journey's Interaction phase is currently a stub;
this fills it. It needs TTS for the Input phase, STT for Speaking, tonal/pronunciation scoring
(which finally feeds the **Listening** and **Speaking** Elo vectors), and a voice agent
constrained to the story's context that feigns misunderstanding (negotiation of meaning). The
30s chi timeout and request/response shape won't fit streaming voice turns, so a streaming or
async transport is part of this phase. Boss Fights (Phase 7) reuse the same conversation engine.
