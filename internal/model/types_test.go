package model

import "testing"

func TestSupportedAgentsContainsKnown(t *testing.T) {
	want := []AgentID{
		AgentCodex, AgentClaudeCode, AgentOpenCode,
		AgentCursor, AgentVSCode, AgentGeminiCLI,
	}
	got := SupportedAgents()
	if len(got) != len(want) {
		t.Fatalf("SupportedAgents length = %d, want %d", len(got), len(want))
	}
	for i, agent := range want {
		if got[i] != agent {
			t.Fatalf("SupportedAgents[%d] = %q, want %q", i, got[i], agent)
		}
	}
}

func TestInstallableAgentsAreSubset(t *testing.T) {
	known := map[AgentID]struct{}{}
	for _, a := range SupportedAgents() {
		known[a] = struct{}{}
	}
	for _, a := range InstallableAgents() {
		if _, ok := known[a]; !ok {
			t.Errorf("installable agent %q not in SupportedAgents", a)
		}
	}
}

func TestIsAgentKnown(t *testing.T) {
	cases := []struct {
		agent AgentID
		want  bool
	}{
		{AgentCodex, true},
		{AgentCursor, true},
		{AgentGeminiCLI, true},
		{AgentID("bogus"), false},
		{AgentID(""), false},
	}
	for _, tc := range cases {
		if got := IsAgentKnown(tc.agent); got != tc.want {
			t.Errorf("IsAgentKnown(%q) = %v, want %v", tc.agent, got, tc.want)
		}
	}
}

func TestIsAgentInstallable(t *testing.T) {
	cases := []struct {
		agent AgentID
		want  bool
	}{
		{AgentCodex, true},
		{AgentClaudeCode, true},
		{AgentOpenCode, true},
		{AgentCursor, false},
		{AgentVSCode, false},
		{AgentGeminiCLI, false},
		{AgentID("bogus"), false},
	}
	for _, tc := range cases {
		if got := IsAgentInstallable(tc.agent); got != tc.want {
			t.Errorf("IsAgentInstallable(%q) = %v, want %v", tc.agent, got, tc.want)
		}
	}
}

func TestSupportedComponentsContainsCore(t *testing.T) {
	required := []ComponentID{ComponentBrain, ComponentFlow}
	got := map[ComponentID]struct{}{}
	for _, c := range SupportedComponents() {
		got[c] = struct{}{}
	}
	for _, want := range required {
		if _, ok := got[want]; !ok {
			t.Errorf("SupportedComponents missing %q", want)
		}
	}
}
