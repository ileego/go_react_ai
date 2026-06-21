package ai

import (
	"fmt"

	"github.com/ileego/go_react_ai/internal/config"
	"github.com/ileego/go_react_ai/pkg/httpx"
)

// ProviderConfig 是创建 Provider 所需的最小配置。
type ProviderConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// NewProvider 根据 config.AIConfig 创建对应的 AIProvider。
func NewProvider(cfg config.AIConfig, client *httpx.Client) (AIProvider, error) {
	apiKey, baseURL, model := resolveConfig(cfg)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for provider %s", cfg.Provider)
	}

	switch cfg.Provider {
	case "openai":
		return NewOpenAIProvider(baseURL, apiKey, model, client), nil
	case "anthropic":
		return NewAnthropicProvider(baseURL, apiKey, model, client), nil
	case "deepseek":
		return NewDeepSeekProvider(baseURL, apiKey, model, client), nil
	case "kimi":
		return NewKimiProvider(baseURL, apiKey, model, client), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.Provider)
	}
}

// resolveConfig 根据 Provider 解析出实际使用的 apiKey、baseURL、model。
func resolveConfig(cfg config.AIConfig) (apiKey, baseURL, model string) {
	switch cfg.Provider {
	case "openai":
		apiKey = firstNonEmpty(cfg.OpenAIAPIKey, cfg.APIKey)
		baseURL = firstNonEmpty(cfg.OpenAIBaseURL, cfg.BaseURL, "https://api.openai.com/v1")
		model = firstNonEmpty(cfg.Model, DefaultModel("openai"))
	case "anthropic":
		apiKey = firstNonEmpty(cfg.AnthropicAPIKey, cfg.APIKey)
		baseURL = firstNonEmpty(cfg.AnthropicBaseURL, cfg.BaseURL, "https://api.anthropic.com")
		model = firstNonEmpty(cfg.Model, DefaultModel("anthropic"))
	case "deepseek":
		apiKey = firstNonEmpty(cfg.DeepSeekAPIKey, cfg.APIKey)
		baseURL = firstNonEmpty(cfg.DeepSeekBaseURL, cfg.BaseURL, "https://api.deepseek.com")
		model = firstNonEmpty(cfg.Model, DefaultModel("deepseek"))
	case "kimi":
		apiKey = firstNonEmpty(cfg.KimiAPIKey, cfg.APIKey)
		baseURL = firstNonEmpty(cfg.KimiBaseURL, cfg.BaseURL, "https://api.moonshot.cn")
		model = firstNonEmpty(cfg.Model, DefaultModel("kimi"))
	}
	return
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
