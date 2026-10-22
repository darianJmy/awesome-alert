// Package provider constructs Eino models without depending on HTTP handlers or global configuration.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

type Config struct{ Provider, BaseURL, APIKey, Model string }

func New(ctx context.Context, cfg Config) (model.ToolCallingChatModel, error) {
	if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("AI key and model are required")
	}
	base, err := BaseURL(cfg.BaseURL, cfg.Provider)
	if err != nil {
		return nil, err
	}
	switch cfg.Provider {
	case "", "anthropic":
		m, err := claude.NewChatModel(ctx, &claude.Config{APIKey: cfg.APIKey, BaseURL: &base, Model: cfg.Model, MaxTokens: 4096, HTTPClient: &http.Client{Timeout: 45 * time.Second}, RequestTimeout: 45 * time.Second})
		if err != nil {
			return nil, err
		}
		return m, nil
	case "newapi", "openai":
		m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{APIKey: cfg.APIKey, BaseURL: base, Model: cfg.Model, Timeout: 45 * time.Second})
		if err != nil {
			return nil, err
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unsupported AI provider")
	}
}

// SDKs append their own endpoint; preserve reverse-proxy prefixes without duplicating /v1.
func BaseURL(base, provider string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid AI base URL")
	}
	path := strings.TrimRight(u.Path, "/")
	switch provider {
	case "", "anthropic":
		path = strings.TrimSuffix(path, "/messages")
		path = strings.TrimSuffix(path, "/v1")
	case "newapi", "openai":
		path = strings.TrimSuffix(path, "/chat/completions")
		if !strings.HasSuffix(path, "/v1") {
			path += "/v1"
		}
	default:
		return "", fmt.Errorf("unsupported AI provider")
	}
	u.Path = path
	u.RawPath = ""
	return u.String(), nil
}
