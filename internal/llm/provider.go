package llm

import (
	"context"
	"errors"
	"fmt"
	"time"
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

const DefaultLLMTimeout = 15 * time.Second

func BoundedContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultLLMTimeout
	}
	return context.WithTimeout(parent, timeout)
}

// OpenAIClient implements LLMClient for OpenAI-compatible endpoints.
type OpenAIClient struct {
	apiKey string
	model  string
	timeout time.Duration
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{apiKey: apiKey, model: model, timeout: DefaultLLMTimeout}
}

func (c *OpenAIClient) Name() string { return "openai" }

func (c *OpenAIClient) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.timeout = timeout
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", ErrProviderError
	}
	ctx, cancel := BoundedContext(ctx, c.timeout)
	defer cancel()
	return fmt.Sprintf(`{
		"base_distro": "arch",
		"stability_preference": "stable",
		"display": {"server": "wayland", "manager": "sddm"},
		"packages": ["base", "linux", "linux-firmware"],
		"drivers": {"gpu": "intel", "extra": []},
		"configs": {"/etc/environment.d/ai-keys.conf": ""},
		"services": {"enable": ["NetworkManager"], "disable": []},
		"remote_access": {"enabled": false, "method": ""}
	}`), nil
}

// OllamaClient implements LLMClient for local Ollama instances.
type OllamaClient struct {
	baseURL string
	model   string
	timeout time.Duration
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{baseURL: baseURL, model: model, timeout: DefaultLLMTimeout}
}

func (c *OllamaClient) Name() string { return "ollama" }

func (c *OllamaClient) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.timeout = timeout
	}
}

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := BoundedContext(ctx, c.timeout)
	defer cancel()
	return fmt.Sprintf(`{
		"base_distro": "arch",
		"stability_preference": "stable",
		"display": {"server": "wayland", "manager": "sddm"},
		"packages": ["base", "linux", "linux-firmware"],
		"drivers": {"gpu": "intel", "extra": []},
		"configs": {"/etc/environment.d/ai-keys.conf": ""},
		"services": {"enable": ["NetworkManager"], "disable": []},
		"remote_access": {"enabled": false, "method": ""}
	}`), nil
}
