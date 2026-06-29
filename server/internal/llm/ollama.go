package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultOllamaBaseURL is the local Ollama daemon address.
const DefaultOllamaBaseURL = "http://localhost:11434"

// OllamaClient talks to a local Ollama daemon's /api/chat endpoint. It is a
// thin, non-streaming adapter: one Request maps to one chat completion.
type OllamaClient struct {
	BaseURL string
	Model   string
	HTTP    *http.Client
}

// NewOllamaClient builds a client for the given model. An empty baseURL falls
// back to DefaultOllamaBaseURL. The HTTP timeout is generous because local
// models (especially larger ones) can take many seconds per generation.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = DefaultOllamaBaseURL
	}
	return &OllamaClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		HTTP:    &http.Client{Timeout: 5 * time.Minute},
	}
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Think    bool            `json:"think"`
	Options  ollamaOptions   `json:"options"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
	Error   string        `json:"error"`
}

// Complete sends the request as a single non-streaming chat call. The system
// prompt becomes a leading system message; thinking is disabled so the reply
// is just the answer text.
func (c *OllamaClient) Complete(ctx context.Context, req Request) (Response, error) {
	msgs := make([]ollamaMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, ollamaMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, ollamaMessage{Role: m.Role, Content: m.Content})
	}

	body, err := json.Marshal(ollamaChatRequest{
		Model:    c.Model,
		Messages: msgs,
		Stream:   false,
		Think:    false,
		Options:  ollamaOptions{Temperature: req.Temperature, NumPredict: req.MaxTokens},
	})
	if err != nil {
		return Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Response{}, fmt.Errorf("ollama status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var out ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Response{}, fmt.Errorf("ollama decode: %w", err)
	}
	if out.Error != "" {
		return Response{}, fmt.Errorf("ollama error: %s", out.Error)
	}
	return Response{Text: out.Message.Content}, nil
}
