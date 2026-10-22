package handlers

import (
	"context"
	"errors"
	"github.com/cloudwego/eino/schema"
	"guanlan/internal/ai/provider"
)

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Compatibility entry point for existing single-turn analysis callers. Transport,
// authentication and protocol conversion are owned by Eino provider adapters.
func callLLM(ctx context.Context, systemPrompt string, messages []anthropicMessage) (string, error) {
	selected, err := resolveAIModel(ctx, 0)
	if err != nil {
		return "", err
	}
	cfg, err := aiModelConfig(selected)
	if err != nil {
		return "", err
	}
	chat, err := provider.New(ctx, cfg)
	if err != nil {
		return "", errors.New("AI 服务配置无效")
	}
	input := make([]*schema.Message, 0, len(messages)+1)
	if systemPrompt != "" {
		input = append(input, schema.SystemMessage(systemPrompt))
	}
	for _, msg := range messages {
		input = append(input, &schema.Message{Role: schema.RoleType(msg.Role), Content: msg.Content})
	}
	result, err := chat.Generate(ctx, input)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", errors.New("AI 模型请求失败")
	}
	if result == nil || result.Content == "" {
		return "", errors.New("AI 返回为空")
	}
	return result.Content, nil
}
