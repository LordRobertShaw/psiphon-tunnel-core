package main

import (
	"testing"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/meshbrowser/agent"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		kind routeKind
		want string
		cap  agent.Capability
	}{
		{name: "https", raw: "https://example.com", kind: routeWeb, want: "https://example.com", cap: agent.CapabilityFetchWeb},
		{name: "bare host", raw: "example.com", kind: routeWeb, want: "https://example.com", cap: agent.CapabilityFetchWeb},
		{name: "nostr", raw: "nostr:note1abc", kind: routeNostr, want: "nostr:note1abc", cap: agent.CapabilityResolveNostr},
		{name: "bitchat", raw: "bitchat:#dr5rs", kind: routeBitChat, want: "bitchat:#dr5rs", cap: agent.CapabilityMeshBitChat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classify(tt.raw)
			if err != nil {
				t.Fatalf("classify() error = %v", err)
			}
			if got.Kind != tt.kind || got.Raw != tt.want || got.Capability != tt.cap {
				t.Fatalf("classify() = %#v, want kind %q raw %q capability %q", got, tt.kind, tt.want, tt.cap)
			}
		})
	}
}

func TestClassifyEmpty(t *testing.T) {
	if _, err := classify("   "); err == nil {
		t.Fatal("classify() error = nil, want error")
	}
}
