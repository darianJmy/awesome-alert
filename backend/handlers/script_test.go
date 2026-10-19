package handlers

import "testing"

func TestBuildScriptAcceptsSupportedRuntime(t *testing.T) {
	script, err := buildScript(CreateScriptReq{Name: "health-check", Type: "python", Content: "print('ok')"})
	if err != nil {
		t.Fatal(err)
	}
	if script.Type != "python" {
		t.Fatalf("type = %q, want python", script.Type)
	}
}

func TestBuildScriptRejectsUnsupportedRuntime(t *testing.T) {
	if _, err := buildScript(CreateScriptReq{Name: "bad", Type: "ruby", Content: "puts 1"}); err == nil {
		t.Fatal("unsupported runtime accepted")
	}
}
