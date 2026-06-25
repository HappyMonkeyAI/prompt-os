package llm

import (
	"context"
	"errors"
)

// LLMClient defines a unified interface for multiple LLM providers.
// This abstraction allows the rest of PromptOS to remain provider-agnostic.
type LLMClient interface {
	// Generate sends a prompt and returns the raw text response.
	Generate(ctx context.Context, prompt string) (string, error)

	// Name returns the provider name (useful for logging / debugging).
	Name() string
}

// Common errors
var (
	ErrEmptyResponse = errors.New("llm: empty response from provider")
	ErrProviderError = errors.New("llm: provider returned an error")
)

// --- Concrete implementations ---

// OpenAIClient implements LLMClient for OpenAI-compatible endpoints.
type OpenAIClient struct {
	apiKey string
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{apiKey: apiKey, model: model}
}

func (c *OpenAIClient) Name() string { return "openai" }

func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	// TODO: Replace with real OpenAI SDK call in Phase 2.2
	if c.apiKey == "" {
		return "", ErrProviderError
	}
	// Placeholder response for now
	return "OpenAI response to: " + prompt, nil
}

// OllamaClient implements LLMClient for local Ollama instances.
type OllamaClient struct {
	baseURL string
	model   string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{baseURL: baseURL, model: model}
}

func (c *OllamaClient) Name() string { return "ollama" }

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	// TODO: Replace with real Ollama HTTP call in Phase 2.2
	return "Ollama response to: " + prompt, nil
}