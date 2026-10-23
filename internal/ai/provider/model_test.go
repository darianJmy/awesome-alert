package provider

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/schema"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNativeToolCallingProviders(t *testing.T) {
	for _, protocol := range []string{"newapi", "anthropic"} {
		for _, suffix := range []string{"", "/v1"} {
			t.Run(protocol+suffix, func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					var body map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					tools, ok := body["tools"].([]interface{})
					if !ok || len(tools) != 1 {
						t.Error("native tools not bound")
					}
					w.Header().Set("Content-Type", "application/json")
					if protocol == "newapi" {
						if r.URL.Path != "/prefix/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer mock" {
							t.Errorf("bad request %s", r.URL.Path)
						}
						if calls == 1 {
							w.Write([]byte(`{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"call1","type":"function","function":{"name":"query","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`))
						} else {
							messages := body["messages"].([]interface{})
							last := messages[len(messages)-1].(map[string]interface{})
							if last["role"] != "tool" || last["tool_call_id"] != "call1" {
								t.Errorf("bad tool result: %+v", last)
							}
							w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"done"},"finish_reason":"stop"}]}`))
						}
					} else {
						if r.URL.Path != "/prefix/v1/messages" || r.Header.Get("x-api-key") != "mock" {
							t.Errorf("bad request %s", r.URL.Path)
						}
						if calls == 1 {
							w.Write([]byte(`{"id":"msg1","type":"message","role":"assistant","model":"test","content":[{"type":"tool_use","id":"call1","name":"query","input":{}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`))
						} else {
							messages := body["messages"].([]interface{})
							last := messages[len(messages)-1].(map[string]interface{})
							content := last["content"].([]interface{})
							block := content[0].(map[string]interface{})
							if block["type"] != "tool_result" || block["tool_use_id"] != "call1" {
								t.Errorf("bad tool result: %+v", block)
							}
							w.Write([]byte(`{"id":"msg2","type":"message","role":"assistant","model":"test","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
						}
					}
				}))
				defer server.Close()
				m, err := New(context.Background(), Config{Provider: protocol, BaseURL: server.URL + "/prefix" + suffix, APIKey: "mock", Model: "test"})
				if err != nil {
					t.Fatal(err)
				}
				bound, err := m.WithTools([]*schema.ToolInfo{{Name: "query", Desc: "query", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{})}})
				if err != nil {
					t.Fatal(err)
				}
				messages := []*schema.Message{schema.UserMessage("query")}
				reply, err := bound.Generate(context.Background(), messages)
				if err != nil {
					t.Fatal(err)
				}
				if len(reply.ToolCalls) != 1 || reply.ToolCalls[0].ID != "call1" {
					t.Fatalf("bad calls: %+v", reply)
				}
				messages = append(messages, reply, schema.ToolMessage("evidence", "call1"))
				final, err := bound.Generate(context.Background(), messages)
				if err != nil || final.Content != "done" {
					t.Fatalf("bad final: %+v %v", final, err)
				}
			})
		}
	}
}
