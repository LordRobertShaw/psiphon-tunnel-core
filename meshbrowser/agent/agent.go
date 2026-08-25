// Package agent provides a small multi-agent coordination layer for MeshBrowser.
//
// The package intentionally keeps agent contracts transport-agnostic so HTTP,
// Psiphon, SCION, Nostr, BitChat, and future transports can compete or
// collaborate without coupling the terminal UI to any one network stack.
package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Capability names a unit of work an agent can perform.
type Capability string

const (
	CapabilityFetchWeb     Capability = "fetch.web"
	CapabilityResolveNostr Capability = "resolve.nostr"
	CapabilityMeshBitChat  Capability = "mesh.bitchat"
	CapabilityRouteSCION   Capability = "route.scion"
)

// Request describes one browser navigation or network-planning task.
type Request struct {
	Target       string
	Capabilities []Capability
}

// Plan is the coordinator's chosen execution plan.
type Plan struct {
	AgentName string
	Score     int
	Steps     []string
}

// Agent is implemented by independent networking or rendering specialists.
type Agent interface {
	Name() string
	Capabilities() []Capability
	Plan(context.Context, Request) (Plan, error)
}

// Coordinator chooses the highest-scoring plan from registered agents.
type Coordinator struct {
	agents []Agent
}

// NewCoordinator returns a coordinator with deterministic agent ordering.
func NewCoordinator(agents ...Agent) *Coordinator {
	ordered := append([]Agent(nil), agents...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Name() < ordered[j].Name() })
	return &Coordinator{agents: ordered}
}

// Plan asks capable agents for a plan and returns the highest-scoring result.
func (c *Coordinator) Plan(ctx context.Context, req Request) (Plan, error) {
	if strings.TrimSpace(req.Target) == "" {
		return Plan{}, fmt.Errorf("empty target")
	}

	var (
		best  Plan
		found bool
	)
	for _, a := range c.agents {
		if !supportsAny(a.Capabilities(), req.Capabilities) {
			continue
		}
		plan, err := a.Plan(ctx, req)
		if err != nil {
			return Plan{}, fmt.Errorf("%s plan: %w", a.Name(), err)
		}
		if !found || plan.Score > best.Score || (plan.Score == best.Score && plan.AgentName < best.AgentName) {
			best = plan
			found = true
		}
	}
	if !found {
		return Plan{}, fmt.Errorf("no agent supports requested capabilities")
	}
	return best, nil
}

func supportsAny(agentCaps, requested []Capability) bool {
	if len(requested) == 0 {
		return true
	}
	for _, have := range agentCaps {
		for _, want := range requested {
			if have == want {
				return true
			}
		}
	}
	return false
}
