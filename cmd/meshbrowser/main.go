// Command meshbrowser is an experimental terminal-first browser prototype that
// sketches how Psiphon networking can be combined with standard HTTP(S), Nostr,
// BitChat-style mesh messaging, and geo-aware relay selection.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/meshbrowser/agent"
)

const defaultTimeout = 10 * time.Second

type routeKind string

const (
	routeWeb     routeKind = "web"
	routeNostr   routeKind = "nostr"
	routeBitChat routeKind = "bitchat"
)

type route struct {
	Kind       routeKind
	Raw        string
	Capability agent.Capability
}

func classify(raw string) (route, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return route{}, fmt.Errorf("empty target")
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "nostr:"):
		return route{Kind: routeNostr, Raw: trimmed, Capability: agent.CapabilityResolveNostr}, nil
	case strings.HasPrefix(lower, "bitchat:"):
		return route{Kind: routeBitChat, Raw: trimmed, Capability: agent.CapabilityMeshBitChat}, nil
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		if _, err := url.ParseRequestURI(trimmed); err != nil {
			return route{}, err
		}
		return route{Kind: routeWeb, Raw: trimmed, Capability: agent.CapabilityFetchWeb}, nil
	default:
		return route{Kind: routeWeb, Raw: "https://" + trimmed, Capability: agent.CapabilityFetchWeb}, nil
	}
}

func render(ctx context.Context, client *http.Client, out io.Writer, target string) error {
	r, err := classify(target)
	if err != nil {
		return err
	}

	plan, err := agent.DefaultCoordinator().Plan(ctx, agent.Request{Target: r.Raw, Capabilities: []agent.Capability{r.Capability, agent.CapabilityRouteSCION}})
	if err != nil {
		return err
	}
	writePlan(out, plan)

	switch r.Kind {
	case routeWeb:
		return renderWeb(ctx, client, out, r.Raw)
	case routeNostr:
		_, err := fmt.Fprintf(out, "Nostr route selected: %s\n\nNext integration: connect a relay pool from rust-nostr or nostr-rest and render matching events.\n", r.Raw)
		return err
	case routeBitChat:
		_, err := fmt.Fprintf(out, "BitChat route selected: %s\n\nNext integration: prefer local mesh transport, then publish/read kind 20000 events through geo-scored Nostr relays.\n", r.Raw)
		return err
	default:
		return fmt.Errorf("unsupported route kind %q", r.Kind)
	}
}

func writePlan(out io.Writer, plan agent.Plan) {
	fmt.Fprintf(out, "Agent plan: %s (score %d)\n", plan.AgentName, plan.Score)
	for _, step := range plan.Steps {
		fmt.Fprintf(out, "- %s\n", step)
	}
	fmt.Fprintln(out)
}

func renderWeb(ctx context.Context, client *http.Client, out io.Writer, target string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "meshbrowser/0 experimental terminal browser")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Fprintf(out, "%s %s\nContent-Type: %s\n\n", resp.Proto, resp.Status, resp.Header.Get("Content-Type"))
	_, err = io.Copy(out, io.LimitReader(resp.Body, 256*1024))
	return err
}

func main() {
	timeout := flag.Duration("timeout", defaultTimeout, "network timeout")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: meshbrowser [--timeout duration] <url|nostr:...|bitchat:...>\n")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := &http.Client{Timeout: *timeout}
	if err := render(ctx, client, os.Stdout, flag.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "meshbrowser: %v\n", err)
		os.Exit(1)
	}
}
