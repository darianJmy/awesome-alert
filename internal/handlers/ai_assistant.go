package handlers

import (
	"context"
	aicore "guanlan/internal/ai"
	"guanlan/internal/ai/provider"
	"guanlan/internal/models"
	"time"
)

// Assemble the assistant at the application boundary. The agent core only sees
// Eino model interfaces, a tool registry and a per-user authorization policy.
func runAIOrchestrator(ctx context.Context, userID, workspaceID uint, question, history, memoryContext string, selectedModel models.AIModel) (string, map[string]interface{}) {
	registry := buildOpsToolRegistry(workspaceID)
	catalog := "[]"
	if authorizeAITool(ctx, userID) == nil {
		catalog = collectDatasourceCatalog(workspaceID)
	} else {
		registry = aicore.NewRegistry()
	}
	registerAssistantMCPTools(ctx, registry, userID)
	modelConfig, configErr := aiModelConfig(selectedModel)
	chatModel, modelErr := provider.New(ctx, modelConfig)
	if configErr != nil || modelErr != nil {
		chatModel = nil
	}
	orchestrator := &aicore.Orchestrator{
		Registry:       registry,
		Model:          chatModel,
		MaxRounds:      8,
		MaxResultBytes: 64 * 1024,
		Timeout:        120 * time.Second,
		Authorize:      aiToolPolicy(userID),
	}
	result := orchestrator.Run(ctx, aicore.RunRequest{
		Question:          question,
		DatasourceCatalog: catalog,
		History:           history,
		Memory:            memoryContext,
		PlatformKnowledge: collectPlatformKnowledge(),
	})
	result.Metadata["model_id"] = selectedModel.ID
	result.Metadata["model_name"] = selectedModel.Name
	if !result.Answered {
		return "", result.Metadata
	}
	return result.Answer, result.Metadata
}
