package ai

import (
	"fmt"
	"strings"
)

// Capability 描述模型的能力。
type Capability struct {
	MaxTokens      int
	SupportsStream bool
	SupportsVision bool
}

// providerModels 保存某 Provider 的模型集合与默认模型。
type providerModels struct {
	Default string
	Models  map[string]Capability
}

var registry = map[string]providerModels{
	"openai": {
		Default: "gpt-4o-mini",
		Models: map[string]Capability{
			"gpt-4o":        {MaxTokens: 128000, SupportsStream: true, SupportsVision: true},
			"gpt-4o-mini":   {MaxTokens: 128000, SupportsStream: true, SupportsVision: true},
			"gpt-4-turbo":   {MaxTokens: 128000, SupportsStream: true, SupportsVision: true},
			"gpt-3.5-turbo": {MaxTokens: 16385, SupportsStream: true, SupportsVision: false},
		},
	},
	"anthropic": {
		Default: "claude-3-haiku-20240307",
		Models: map[string]Capability{
			"claude-3-opus-20240229":    {MaxTokens: 200000, SupportsStream: true, SupportsVision: true},
			"claude-3-sonnet-20240229":  {MaxTokens: 200000, SupportsStream: true, SupportsVision: true},
			"claude-3-haiku-20240307":   {MaxTokens: 200000, SupportsStream: true, SupportsVision: true},
			"claude-3-5-sonnet-20240620": {MaxTokens: 200000, SupportsStream: true, SupportsVision: true},
		},
	},
	"deepseek": {
		Default: "deepseek-chat",
		Models: map[string]Capability{
			"deepseek-chat":     {MaxTokens: 64000, SupportsStream: true, SupportsVision: false},
			"deepseek-reasoner": {MaxTokens: 64000, SupportsStream: true, SupportsVision: false},
		},
	},
	"kimi": {
		Default: "moonshot-v1-8k",
		Models: map[string]Capability{
			"moonshot-v1-8k":   {MaxTokens: 8192, SupportsStream: true, SupportsVision: false},
			"moonshot-v1-32k":  {MaxTokens: 32768, SupportsStream: true, SupportsVision: false},
			"moonshot-v1-128k": {MaxTokens: 128000, SupportsStream: true, SupportsVision: false},
		},
	},
}

// SupportedProviders 返回所有支持的 Provider 名称。
func SupportedProviders() []string {
	providers := make([]string, 0, len(registry))
	for k := range registry {
		providers = append(providers, k)
	}
	return providers
}

// DefaultModel 返回指定 Provider 的默认模型。
func DefaultModel(provider string) string {
	pm, ok := registry[strings.ToLower(provider)]
	if !ok {
		return ""
	}
	return pm.Default
}

// ValidateModel 校验模型是否被 Provider 支持。
func ValidateModel(provider, model string) error {
	pm, ok := registry[strings.ToLower(provider)]
	if !ok {
		return fmt.Errorf("unsupported provider: %s", provider)
	}
	if _, ok := pm.Models[model]; !ok {
		return fmt.Errorf("unsupported model %q for provider %q", model, provider)
	}
	return nil
}

// GetCapability 返回指定 Provider+Model 的能力信息。
func GetCapability(provider, model string) (Capability, bool) {
	pm, ok := registry[strings.ToLower(provider)]
	if !ok {
		return Capability{}, false
	}
	cap, ok := pm.Models[model]
	return cap, ok
}

// ListModels 返回某 Provider 支持的所有模型。
func ListModels(provider string) []string {
	pm, ok := registry[strings.ToLower(provider)]
	if !ok {
		return nil
	}
	models := make([]string, 0, len(pm.Models))
	for k := range pm.Models {
		models = append(models, k)
	}
	return models
}

// SupportsStream 判断某 Provider+Model 是否支持流式。
func SupportsStream(provider, model string) bool {
	cap, ok := GetCapability(provider, model)
	return ok && cap.SupportsStream
}
