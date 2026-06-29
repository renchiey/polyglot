package lexaudit

import (
	"fmt"
	"sync"
)

// builder constructs an Auditor on demand.
type builder func() (*Auditor, error)

// Registry resolves language codes to lazily-built, cached Auditors. It is
// safe for concurrent use.
type Registry struct {
	mu       sync.Mutex
	builders map[string]builder
	cache    map[string]*Auditor
}

// NewRegistry builds the registry with all supported languages registered.
// Adding a language is a single entry here plus its builder.
func NewRegistry() *Registry {
	r := &Registry{
		builders: map[string]builder{},
		cache:    map[string]*Auditor{},
	}

	for _, code := range []string{"zh", "cmn", "mandarin"} {
		r.builders[code] = newMandarinAuditor
	}

	return r
}

// Supported reports whether a language code is registered.
func (r *Registry) Supported(code string) bool {
	_, ok := r.builders[code]
	return ok
}

// Get returns the Auditor for code, building and caching it on first use.
// An unsupported code returns an error distinct from a build failure; callers
// can use Supported to tell them apart.
func (r *Registry) Get(code string) (*Auditor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if a, ok := r.cache[code]; ok {
		return a, nil
	}
	build, ok := r.builders[code]
	if !ok {
		return nil, fmt.Errorf("unsupported language %q", code)
	}
	a, err := build()
	if err != nil {
		return nil, err
	}
	r.cache[code] = a
	return a, nil
}

// Warm eagerly builds the default language so the first request avoids the
// initial dictionary load and data download.
func (r *Registry) Warm() error {
	_, err := r.Get("zh")
	return err
}
