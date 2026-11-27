// Package llm is the provider-agnostic entry point for LLM completions. Business code depends on
// the Client interface only; the concrete provider lives in its own file. Keys and endpoints come
// from the application config (global.GVA_CONFIG.LLM) and are never hard-coded.
package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Usage token accounting as reported by the provider.
type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// Client one text completion call. Implementations must respect ctx deadlines.
type Client interface {
	Complete(ctx context.Context, system string, user string) (text string, usage Usage, err error)
}

// Config everything a provider needs. Provider selects the implementation; the remaining fields
// are provider-specific but share the OpenAI-compatible shape for the shipped provider.
type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	Timeout     time.Duration
	MaxRetries  int
	Temperature float64
}

var (
	// ErrNotConfigured the provider or API key is missing: the caller should fall back to rules.
	ErrNotConfigured = errors.New("llm: provider is not configured")
	// ErrEmptyResponse the provider answered successfully but without content.
	ErrEmptyResponse = errors.New("llm: provider returned an empty completion")
)

// APIError a non-2xx answer from the provider.
type APIError struct {
	Provider   string
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm: %s returned status %d: %s", e.Provider, e.StatusCode, e.Message)
}

// Retryable 429 and 5xx are worth retrying; other 4xx are terminal.
func (e *APIError) Retryable() bool {
	return e.StatusCode == 429 || e.StatusCode >= 500
}

// ConfigFrom builds a Config from loose config values (kept provider-agnostic for callers).
func ConfigFrom(provider, model, apiKey, baseURL string, timeoutSeconds, maxRetries int, temperature float64) Config {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}
	if maxRetries < 0 {
		maxRetries = 0
	}
	return Config{
		Provider: provider, Model: model, APIKey: apiKey, BaseURL: baseURL,
		Timeout: time.Duration(timeoutSeconds) * time.Second, MaxRetries: maxRetries, Temperature: temperature,
	}
}

// NewClient is the provider factory. It supports OpenAI-compatible chat-completions APIs
// (OpenAI, Azure-OpenAI gateways, OpenRouter, Groq, Together, vLLM/Ollama proxies).
func NewClient(cfg Config) (Client, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "none", "disabled":
		return nil, ErrNotConfigured
	case "openai", "azure-openai", "openrouter", "groq", "together", "ollama", "custom":
		return newOpenAIClient(cfg), nil
	default:
		return nil, fmt.Errorf("llm: unsupported provider %q", cfg.Provider)
	}
}

// withRetry runs fn up to maxRetries extra times with exponential backoff on retryable failures,
// honouring the context deadline. Non-retryable provider errors return immediately.
func withRetry(ctx context.Context, maxRetries int, fn func() (string, Usage, error)) (string, Usage, error) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return "", Usage{}, ctx.Err()
			case <-time.After(wait):
			}
		}
		text, usage, err := fn()
		if err == nil {
			return text, usage, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return "", Usage{}, ctx.Err()
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) && !apiErr.Retryable() {
			return "", Usage{}, err
		}
	}
	return "", Usage{}, lastErr
}
