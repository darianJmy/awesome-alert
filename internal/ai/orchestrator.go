package ai

import (
	"context"
	"github.com/cloudwego/eino/components/model"
	"strings"
	"time"
)

type Orchestrator struct {
	Registry       *Registry
	Model          model.ToolCallingChatModel
	MaxRounds      int
	MaxResultBytes int
	Timeout        time.Duration
	Authorize      func(context.Context, ToolSpec, ToolCall) error
}

type RunRequest struct {
	Question          string
	DatasourceCatalog string
	History           string
	Memory            string
	PlatformKnowledge string
}

type RunResponse struct {
	Answered bool
	Answer   string
	Metadata map[string]interface{}
}

func (o *Orchestrator) systemPrompt() string {
	return `你是观澜运维平台的智能运维调度核心。
你必须使用中文。
使用提供的工具进行查询，最终回答使用清晰的中文，可使用 Markdown。
规则：
- 如果问题是普通知识、概念解释、闲聊，直接回答。
- 如果问题需要查询平台真实数据，必须调用工具。
- 当前阶段仅支持只读查询；涉及创建、修改、删除或批量变更时说明暂不支持，不得声称已执行。
- 工具返回失败或没有证据时如实说明，不能编造查询结果。
- 你看不到也不能要求用户提供密钥、token、AK/SK、URL、header。
- 你只能使用工具清单中的工具名和参数。
- 告警、规则、拨测和自动化任务目录查询属于平台级范围，不按当前数据源分组筛选；请明确区分平台查询与分组内指标查询。
- 当前指标数据源支持 Prometheus 和 VictoriaMetrics；不能直接查询 Kubernetes 资源、云资源、数据库或日志服务。
- 工具失败时，根据错误调整参数或给出可执行解释。`
}

func (o *Orchestrator) initialPrompt(req RunRequest) string {
	var sb strings.Builder
	sb.WriteString("用户问题:\n")
	sb.WriteString(req.Question)
	sb.WriteString("\n\n可用数据源目录（只含 id/name/type/capability，不含密钥和真实连接细节）:\n")
	sb.WriteString(req.DatasourceCatalog)
	if strings.TrimSpace(req.PlatformKnowledge) != "" {
		sb.WriteString("\n\n平台知识:\n")
		sb.WriteString(req.PlatformKnowledge)
	}
	if strings.TrimSpace(req.History) != "" {
		sb.WriteString("\n\n最近对话:\n")
		sb.WriteString(req.History)
	}
	if strings.TrimSpace(req.Memory) != "" {
		sb.WriteString("\n\n长期记忆:\n")
		sb.WriteString(req.Memory)
	}
	return sb.String()
}
