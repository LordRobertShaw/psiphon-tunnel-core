package agent

import (
	"context"
	"fmt"
	"strings"
)

// StaticAgent is a lightweight deterministic agent used until each transport is
// backed by its native upstream implementation.
type StaticAgent struct {
	AgentName string
	Caps      []Capability
	BaseScore int
	Steps     []string
}

func (a StaticAgent) Name() string { return a.AgentName }

func (a StaticAgent) Capabilities() []Capability { return append([]Capability(nil), a.Caps...) }

func (a StaticAgent) Plan(_ context.Context, req Request) (Plan, error) {
	steps := make([]string, 0, len(a.Steps)+1)
	steps = append(steps, fmt.Sprintf("target: %s", strings.TrimSpace(req.Target)))
	steps = append(steps, a.Steps...)
	return Plan{AgentName: a.AgentName, Score: a.BaseScore, Steps: steps}, nil
}

// DefaultCoordinator wires the foundational agents MeshBrowser can consult.
func DefaultCoordinator() *Coordinator {
	return NewCoordinator(
		StaticAgent{
			AgentName: "web-agent",
			Caps:      []Capability{CapabilityFetchWeb},
			BaseScore: 50,
			Steps: []string{
				"fetch via standard HTTP(S) client",
				"preserve Psiphon tunnel hooks behind the netstack interface",
			},
		},
		StaticAgent{
			AgentName: "nostr-agent",
			Caps:      []Capability{CapabilityResolveNostr},
			BaseScore: 70,
			Steps: []string{
				"resolve Nostr URI into events or profiles",
				"query geo-scored relays directly or through nostr-rest",
			},
		},
		StaticAgent{
			AgentName: "bitchat-agent",
			Caps:      []Capability{CapabilityMeshBitChat},
			BaseScore: 80,
			Steps: []string{
				"try local BitChat mesh discovery first",
				"fallback to kind 20000 events on Nostr relays",
			},
		},
		StaticAgent{
			AgentName: "scion-agent",
			Caps:      []Capability{CapabilityRouteSCION},
			BaseScore: 60,
			Steps: []string{
				"evaluate SCION path-aware routes before default internet egress",
				"prefer authenticated multi-path routes when available",
			},
		},
	)
}
