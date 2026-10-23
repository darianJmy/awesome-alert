package ai

import (
	"context"
	"fmt"
	"sort"
)

type ToolHandler func(ctx context.Context, call ToolCall) ToolResult

type Tool struct {
	Spec    ToolSpec
	Handler ToolHandler
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(spec ToolSpec, handler ToolHandler) error {
	if spec.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if handler == nil {
		return fmt.Errorf("tool handler is required: %s", spec.Name)
	}
	if _, exists := r.tools[spec.Name]; exists {
		return fmt.Errorf("tool already registered: %s", spec.Name)
	}
	r.tools[spec.Name] = Tool{Spec: spec, Handler: handler}
	return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Specs() []ToolSpec {
	specs := make([]ToolSpec, 0, len(r.tools))
	for _, tool := range r.tools {
		specs = append(specs, tool.Spec)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs
}
