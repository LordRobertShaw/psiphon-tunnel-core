package agent

import (
	"context"
	"testing"
)

func TestCoordinatorSelectsHighestScoringCapableAgent(t *testing.T) {
	coordinator := NewCoordinator(
		StaticAgent{AgentName: "low", Caps: []Capability{CapabilityFetchWeb}, BaseScore: 1},
		StaticAgent{AgentName: "high", Caps: []Capability{CapabilityFetchWeb}, BaseScore: 10},
	)

	plan, err := coordinator.Plan(context.Background(), Request{Target: "https://example.com", Capabilities: []Capability{CapabilityFetchWeb}})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if plan.AgentName != "high" {
		t.Fatalf("Plan() AgentName = %q, want high", plan.AgentName)
	}
}

func TestCoordinatorNoCapableAgent(t *testing.T) {
	coordinator := NewCoordinator(StaticAgent{AgentName: "web", Caps: []Capability{CapabilityFetchWeb}, BaseScore: 1})
	if _, err := coordinator.Plan(context.Background(), Request{Target: "nostr:note1abc", Capabilities: []Capability{CapabilityResolveNostr}}); err == nil {
		t.Fatal("Plan() error = nil, want error")
	}
}
