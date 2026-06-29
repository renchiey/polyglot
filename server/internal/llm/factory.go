package llm

import "fmt"

// New constructs a Client for the named provider. model and baseURL are used
// by providers that need them (e.g. ollama); they are ignored by "mock".
//
// "mock" (the default for empty input) returns a MockClient with a harmless
// canned reply so the server boots and /generate works without credentials.
// "ollama" targets a local Ollama daemon. "anthropic" and "openai" are
// recognized but not yet implemented, returning a clear error rather than a
// silent fallback so the seam is obvious when those adapters land.
func New(provider, model, baseURL string) (Client, error) {
	switch provider {
	case "", "mock":
		return NewMockClient("你好。"), nil
	case "ollama":
		if model == "" {
			return nil, fmt.Errorf("ollama provider requires LLM_MODEL")
		}
		return NewOllamaClient(baseURL, model), nil
	case "anthropic", "openai":
		return nil, fmt.Errorf("llm provider %q not yet implemented", provider)
	default:
		return nil, fmt.Errorf("unknown llm provider %q", provider)
	}
}
