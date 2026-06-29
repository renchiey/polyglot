package llm

import (
	"context"
	"sync"
)

// MockClient is a deterministic Client for local runs and tests. It returns
// scripted replies in order, falling back to Default once the script is
// exhausted. It records every Request it receives for assertions.
//
// Safe for concurrent use, though the pipeline calls it serially.
type MockClient struct {
	// Default is returned once Script is exhausted (or when Script is empty).
	Default string

	mu     sync.Mutex
	script []string
	calls  []Request
}

// NewMockClient builds a MockClient that replays replies in order, then
// returns def for every call thereafter.
func NewMockClient(def string, replies ...string) *MockClient {
	return &MockClient{Default: def, script: replies}
}

// Complete returns the next scripted reply, or Default when the script runs out.
func (m *MockClient) Complete(_ context.Context, req Request) (Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, req)
	if len(m.script) > 0 {
		reply := m.script[0]
		m.script = m.script[1:]
		return Response{Text: reply}, nil
	}
	return Response{Text: m.Default}, nil
}

// Calls returns a copy of the requests received so far, in call order.
func (m *MockClient) Calls() []Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Request(nil), m.calls...)
}
