package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// openAIClient talks to any OpenAI-compatible /chat/completions endpoint.
type openAIClient struct {
	cfg  Config
	http *http.Client
}

func newOpenAIClient(cfg Config) *openAIClient {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultOpenAIBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	return &openAIClient{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openAIClient) Complete(ctx context.Context, system string, user string) (string, Usage, error) {
	if strings.TrimSpace(c.cfg.APIKey) == "" {
		return "", Usage{}, ErrNotConfigured
	}

	body := chatRequest{
		Model:       c.cfg.Model,
		Temperature: c.cfg.Temperature,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: &struct {
			Type string `json:"type"`
		}{Type: "json_object"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", Usage{}, err
	}

	return withRetry(ctx, c.cfg.MaxRetries, func() (string, Usage, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			return "", Usage{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

		resp, err := c.http.Do(req)
		if err != nil {
			return "", Usage{}, err
		}
		defer resp.Body.Close()
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return "", Usage{}, readErr
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := strings.TrimSpace(string(data))
			var parsed chatResponse
			if json.Unmarshal(data, &parsed) == nil && parsed.Error != nil && parsed.Error.Message != "" {
				msg = parsed.Error.Message
			}
			if len(msg) > 500 {
				msg = msg[:500]
			}
			return "", Usage{}, &APIError{Provider: c.cfg.Provider, StatusCode: resp.StatusCode, Message: msg}
		}

		var out chatResponse
		if err := json.Unmarshal(data, &out); err != nil {
			return "", Usage{}, err
		}
		if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
			return "", Usage{}, ErrEmptyResponse
		}
		return out.Choices[0].Message.Content, Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		}, nil
	})
}
