package handlers

import (
	"reflect"
	"testing"
)

func TestAgentVersionAtLeast(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    bool
	}{
		{name: "minimum", current: "0.1.1", want: true},
		{name: "newer patch", current: "0.1.2", want: true},
		{name: "newer minor", current: "0.2.0", want: true},
		{name: "v prefix", current: "v0.1.1", want: true},
		{name: "old", current: "0.1.0", want: false},
		{name: "unknown", current: "", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := agentVersionAtLeast(test.current, minimumAgentExecVersion); got != test.want {
				t.Fatalf("agentVersionAtLeast(%q, %q) = %v, want %v", test.current, minimumAgentExecVersion, got, test.want)
			}
		})
	}
}

func TestInitialAgentCapabilities(t *testing.T) {
	tests := []struct {
		name           string
		deploymentType string
		dockerImage    string
		want           []string
	}{
		{name: "host agent", deploymentType: "agent", want: []string{"runtime:host"}},
		{name: "docker agent", deploymentType: "docker", dockerImage: "registry.example.com/aiops/python:3.12", want: []string{"runtime:docker", "docker_image:registry.example.com/aiops/python:3.12"}},
		{name: "unknown defaults to host", deploymentType: "other", want: []string{"runtime:host"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := initialAgentCapabilities(test.deploymentType, test.dockerImage); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("initialAgentCapabilities() = %#v, want %#v", got, test.want)
			}
		})
	}
}
