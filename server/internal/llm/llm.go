// Package llm is a provider-agnostic interface to large language models.
//
// The rest of the backend depends only on the Client interface and the
// Request/Response/Message types here, never on a concrete provider. A
// deterministic MockClient lets the generation pipeline run and be tested
// without network access or API keys; real adapters (Anthropic, OpenAI) plug
// in behind New without touching callers.
package llm

import "context"

// Role values for a Message. Mirrors the Anthropic Messages / OpenAI Chat
// split so an adapter can map them without translation.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is a single turn in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is a single completion call. System is the system prompt (kept
// separate from Messages to match the Anthropic Messages API); Messages is the
// conversation so far, ending with the user turn to respond to.
type Request struct {
	System      string    `json:"system,omitempty"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Response is the model's reply.
type Response struct {
	Text string `json:"text"`
}

// Client completes a request against some model provider.
type Client interface {
	Complete(ctx context.Context, req Request) (Response, error)
}
