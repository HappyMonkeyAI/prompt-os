package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient defines a unified interface for multiple LLM providers.
type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Name() string
}

var (
	ErrEmptyResponse = fmt.Errorf("llm: empty response from provider")
	ErrProviderError = fmt.Errorf("llm: provider returned an error")
)

const DefaultLLMTimeout = 60 * time.Second

func BoundedContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultLLMTimeout
	}
	return context.WithTimeout(parent, timeout)
}

// ---- shared request/response types ----------------------------------------

type openAIChatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ---- OpenAI ----------------------------------------------------------------

// OpenAIClient calls the OpenAI chat completions API.
type OpenAIClient struct {
	apiKey  string
	model   string
	timeout time.Duration
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{apiKey: apiKey, model: model, timeout: DefaultLLMTimeout}
}

func (c *OpenAIClient) Name() string { return "openai" }

func (c *OpenAIClient) SetTimeout(d time.Duration) {
	if d > 0 {
		c.timeout = d
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("%w: no API key set", ErrProviderError)
	}
	ctx, cancel := BoundedContext(ctx, c.timeout)
	defer cancel()
	return openAIChat(ctx, "https://api.openai.com/v1/chat/completions", c.apiKey, c.model, "", prompt)
}

// ---- OpenRouter ------------------------------------------------------------

// OpenRouterClient calls OpenRouter's OpenAI-compatible API.
// OpenRouter supports hundreds of models via a single endpoint.
type OpenRouterClient struct {
	apiKey  string
	model   string
	timeout time.Duration
}

func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	if model == "" {
		// A capable, cost-effective default available on OpenRouter
		model = "openai/gpt-4o-mini"
	}
	return &OpenRouterClient{apiKey: apiKey, model: model, timeout: DefaultLLMTimeout}
}

func (c *OpenRouterClient) Name() string { return "openrouter" }

func (c *OpenRouterClient) SetTimeout(d time.Duration) {
	if d > 0 {
		c.timeout = d
	}
}

func (c *OpenRouterClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("%w: no API key set", ErrProviderError)
	}
	ctx, cancel := BoundedContext(ctx, c.timeout)
	defer cancel()
	// OpenRouter is OpenAI-compatible; pass a Referer so it identifies the app
	return openAIChat(ctx, "https://openrouter.ai/api/v1/chat/completions", c.apiKey, c.model, "https://github.com/HappyMonkeyAI/prompt-os", prompt)
}

// ---- Ollama ----------------------------------------------------------------

// OllamaClient calls a local Ollama instance.
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

func (c *OllamaClient) SetTimeout(d time.Duration) {
	if d > 0 {
		c.timeout = d
	}
}

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := BoundedContext(ctx, c.timeout)
	defer cancel()

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderError, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d: %s", ErrProviderError, resp.StatusCode, string(raw))
	}

	var result ollamaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.Response == "" {
		return "", ErrEmptyResponse
	}
	return result.Response, nil
}

// ---- NewClientFromProvider -------------------------------------------------

// NewClientFromProvider constructs an LLMClient from a provider name and API key.
// For ollama the apiKey is ignored (no auth needed).
func NewClientFromProvider(provider, apiKey string) (LLMClient, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient(apiKey, ""), nil
	case "openrouter":
		return NewOpenRouterClient(apiKey, ""), nil
	case "ollama":
		return NewOllamaClient("", ""), nil
	default:
		return nil, fmt.Errorf("unknown provider: %q", provider)
	}
}

// ---- shared OpenAI-compatible helper ---------------------------------------

func openAIChat(ctx context.Context, url, apiKey, model, referer, prompt string) (string, error) {
	reqBody := openAIChatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: prompt},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if referer != "" {
		req.Header.Set("HTTP-Referer", referer)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderError, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d: %s", ErrProviderError, resp.StatusCode, string(raw))
	}

	var result openAIChatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", fmt.Errorf("%w: %s", ErrProviderError, result.Error.Message)
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", ErrEmptyResponse
	}
	return result.Choices[0].Message.Content, nil
}
