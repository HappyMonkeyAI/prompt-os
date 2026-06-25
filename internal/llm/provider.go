package llm

import (
	"context"
	"errors"
	"fmt"
)

// LLMClient defines a unified interface for multiple LLM providers.
type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Name() string
}

var (
	ErrEmptyResponse = errors.New("llm: empty response from provider")
	ErrProviderError = errors.New("llm: provider returned an error")
)

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
	if c.apiKey == "" {
		return "", ErrProviderError
	}
	// Return valid dummy JSON for development/testing
	return fmt.Sprintf(`{
		"base_distro": "arch",
		"stability_preference": "stable",
		"display": {"server": "wayland", "manager": "sddm"},
		"packages": ["base", "linux", "linux-firmware"],
		"drivers": {"gpu": "intel", "extra": []},
		"configs": {"/etc/environment.d/ai-keys.conf": "OPENAI_API_KEY=sk-..."},
		"services": {"enable": ["NetworkManager"], "disable": []},
		"remote_access": {"enabled": false, "method": ""}
	}`), nil
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
	// Return valid dummy JSON for development/testing
	return fmt.Sprintf(`{
		"base_distro": "arch",
		"stability_preference": "stable",
		"display": {"server": "wayland", "manager": "sddm"},
		"packages": ["base", "linux", "linux-firmware"],
		"drivers": {"gpu": "intel", "extra": []},
		"configs": {"/etc/environment.d/ai-keys.conf": "OPENAI_API_KEY=sk-..."},
		"services": {"enable": ["NetworkManager"], "disable": []},
		"remote_access": {"enabled": false, "method": ""}
	}`), nil
}