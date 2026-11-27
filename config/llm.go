package config

// LLM provider settings for the recommendation engine. The key is read from config/env only and
// must never be committed or logged. Provider names map to utils/llm implementations
// (openai | azure-openai | openrouter | groq | together | ollama | custom).
type LLM struct {
	Provider       string  `mapstructure:"provider" json:"provider" yaml:"provider"`
	Model          string  `mapstructure:"model" json:"model" yaml:"model"`
	APIKey         string  `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL        string  `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	TimeoutSeconds int     `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"`
	MaxRetries     int     `mapstructure:"max-retries" json:"max-retries" yaml:"max-retries"`
	Temperature    float64 `mapstructure:"temperature" json:"temperature" yaml:"temperature"`
}
