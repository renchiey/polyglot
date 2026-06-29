package server

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/renchieyang/polyglot/server/internal/auth"
	"github.com/renchieyang/polyglot/server/internal/config"
	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/dict"
	"github.com/renchieyang/polyglot/server/internal/generate"
	"github.com/renchieyang/polyglot/server/internal/handlers"
	"github.com/renchieyang/polyglot/server/internal/journey"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
	"github.com/renchieyang/polyglot/server/internal/progress"
	"github.com/renchieyang/polyglot/server/internal/srs"
	"github.com/renchieyang/polyglot/server/internal/tts"
)

// New builds the HTTP router with all routes and middleware wired up.
func New(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	queries := gen.New(pool)
	authHandler := &handlers.AuthHandler{Queries: queries, JWTSecret: cfg.JWTSecret}

	registry := lexaudit.NewRegistry()
	// Warm the default lexicon in the background so the first audit request
	// doesn't pay the dictionary build. Failures are logged; the /audit route
	// reports 503 until the lexicon is available.
	go func() {
		if err := registry.Warm(); err != nil {
			log.Printf("lexaudit warm: %v", err)
		}
	}()
	auditHandler := &handlers.AuditHandler{Registry: registry}

	// Build the LLM client for the generation pipeline. An unimplemented or
	// misconfigured provider falls back to the deterministic mock so the server
	// still boots; /generate then returns canned content rather than failing.
	llmClient, err := llm.New(cfg.LLMProvider, cfg.LLMModel, cfg.LLMBaseURL)
	if err != nil {
		log.Printf("llm provider %q: %v; falling back to mock", cfg.LLMProvider, err)
		llmClient, _ = llm.New("mock", "", "")
	}
	tracker := progress.NewTracker(queries, registry)

	pipeline := generate.NewPipeline(llmClient, generate.DefaultMaxRounds)
	generateHandler := &handlers.GenerateHandler{Pipeline: pipeline, Registry: registry, Queries: queries, Tracker: tracker}
	progressHandler := &handlers.ProgressHandler{Tracker: tracker, Registry: registry}

	scheduler := srs.NewScheduler()
	wordsHandler := &handlers.WordsHandler{Queries: queries, Scheduler: scheduler}
	cardsHandler := &handlers.CardsHandler{Queries: queries, Scheduler: scheduler, Tracker: tracker}

	// The dictionary is embedded, so loading is fast and offline; on the
	// unlikely parse failure the dict-backed routes report 503 (nil Dict).
	dictionary, err := dict.Load()
	if err != nil {
		log.Printf("dict load: %v", err)
	}
	lookupHandler := &handlers.LookupHandler{Dict: dictionary}
	segmentHandler := &handlers.SegmentHandler{Registry: registry, Dict: dictionary}
	recallHandler := &handlers.RecallHandler{Pipeline: pipeline, Registry: registry, Queries: queries, Dict: dictionary, Tracker: tracker}
	assessHandler := &handlers.AssessHandler{LLM: llmClient, Registry: registry, Tracker: tracker}
	journeyHandler := &handlers.JourneyHandler{
		Builder:  journey.NewBuilder(pipeline, dictionary),
		Registry: registry,
		Queries:  queries,
		Tracker:  tracker,
	}

	// TTS auto-discovers a voice in VoicesDir and a piper command. Unavailable
	// => the route 503s and the web client falls back to browser speech.
	ttsHandler := &handlers.TTSHandler{Synth: tts.New(cfg.PiperBin, cfg.PiperVoice, cfg.VoicesDir)}
	if ttsHandler.Synth.Available() {
		log.Printf("tts: piper voice ready")
	} else {
		log.Printf("tts: no piper voice (run `make tts-setup`); using browser fallback")
	}

	r.Get("/health", handlers.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Get("/me", authHandler.Me)
		r.Get("/progress", progressHandler.Progress)
		r.Post("/audit", auditHandler.Audit)
		r.Post("/audit/batch", auditHandler.AuditBatch)
		r.Post("/generate", generateHandler.Generate)

		r.Route("/words", func(r chi.Router) {
			r.Post("/", wordsHandler.Create)
			r.Get("/", wordsHandler.List)
			r.Get("/{id}", wordsHandler.Get)
			r.Put("/{id}", wordsHandler.Update)
			r.Delete("/{id}", wordsHandler.Delete)
		})

		r.Get("/cards/due", cardsHandler.Due)
		r.Post("/cards/review", cardsHandler.Review)

		r.Get("/lookup", lookupHandler.Lookup)
		r.Post("/segment", segmentHandler.Segment)
		r.Post("/recall", recallHandler.Recall)
		r.Post("/assess/translation", assessHandler.Translation)
		r.Post("/tts", ttsHandler.Speak)

		r.Route("/journey", func(r chi.Router) {
			r.Post("/start", journeyHandler.Start)
			r.Get("/{id}", journeyHandler.Get)
		})
	})

	return r
}
