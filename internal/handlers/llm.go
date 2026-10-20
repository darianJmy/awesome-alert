package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"guanlan/internal/config"
)

const anthropicVersion = "2023-06-01"

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callAnthropic(ctx context.Context, systemPrompt string, messages []anthropicMessage) (string, error) {
	if strings.TrimSpace(config.AppConfig.AnthropicBaseURL) == "" {
		return "", errors.New("AI 服务地址未配置，请设置 ai.base_url 或 APP_AI_BASE_URL")
	}
	if strings.TrimSpace(config.AppConfig.AnthropicAPIKey) == "" {
		return "", errors.New("ANTHROPIC_API_KEY 未配置")
	}

	reqBody := anthropicRequest{
		Model:     config.AppConfig.AnthropicModel,
		MaxTokens: 1200,
		System:    systemPrompt,
		Messages:  messages,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(config.AppConfig.AnthropicBaseURL, "/") + "/v1/messages"
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			if err := waitForLLMRetry(ctx, time.Duration(attempt*attempt)*500*time.Millisecond); err != nil {
				return "", err
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		httpReq.Header.Set("content-type", "application/json")
		httpReq.Header.Set("x-api-key", config.AppConfig.AnthropicAPIKey)
		httpReq.Header.Set("anthropic-version", anthropicVersion)

		client := &http.Client{Timeout: 45 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return "", readErr
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("LLM 服务暂时不可用: HTTP %d", resp.StatusCode)
			if retryAfter := retryAfterDelay(resp.Header.Get("Retry-After")); retryAfter > 0 {
				if attempt < 2 {
					if err := waitForLLMRetry(ctx, retryAfter); err != nil {
						return "", err
					}
				}
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("LLM 请求失败: HTTP %d: %s", resp.StatusCode, truncateStr(string(respBody), 300))
		}

		var parsed anthropicResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return "", err
		}
		if parsed.Error != nil {
			return "", fmt.Errorf("LLM 返回错误: %s", parsed.Error.Message)
		}

		var sb strings.Builder
		for _, item := range parsed.Content {
			if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(item.Text)
			}
		}
		if sb.Len() == 0 {
			return "", errors.New("LLM 返回为空")
		}
		return sb.String(), nil
	}

	return "", lastErr
}

func retryAfterDelay(value string) time.Duration {
	if value == "" {
		return 0
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func waitForLLMRetry(ctx context.Context, delay time.Duration) error {
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
