# MeshBrowser architecture

MeshBrowser is an experimental terminal-first browser concept for combining standard web networking with Nostr relay discovery, BitChat-style mesh messaging, geo-aware relay selection, and SCION-style path-aware routing under a small multi-agent coordinator.

## Upstream components to integrate

- `georelays` supplies the relay discovery and geolocation data pipeline. Its output can seed nearby or latency-friendly Nostr relays for browser metadata, chat rooms, and fallback routing.
- `bitchat` supplies the dual-transport product model: local mesh first, internet Nostr fallback, IRC-style commands, location channels, and kind `20000` BitChat relay compatibility.
- `nostr-rest` supplies a REST bridge pattern for environments where direct WebSocket relay access is undesirable or unavailable.
- `nostr` supplies the long-term Rust protocol implementation and SDK surface for Nostr events, NIPs, relay pools, storage, and signing.
- `scion` supplies the path-aware networking foundation for authenticated, multi-path routing decisions before falling back to regular internet egress.

## Browser goals

1. Keep the terminal UI fast: fetch content concurrently, stream text output, and avoid heavy rendering work by default.
2. Support normal URLs through HTTP and HTTPS first.
3. Treat `nostr:` and `bitchat:` URIs as first-class navigation targets instead of external side channels.
4. Route BitChat messages by preference order: local mesh transport, then Nostr relay fallback, then optional REST bridge.
5. Use geo-relay data to bias relay selection toward nearby, healthy, BitChat-compatible relays.
6. Preserve Psiphon tunnel compatibility by keeping transport implementations behind small interfaces.
7. Use a multi-agent coordinator so web, Nostr, BitChat, geo-relay, SCION, and rendering specialists can independently score routing plans.

## Proposed package boundaries

- `cmd/meshbrowser`: terminal command entry point and prototype navigation loop.
- `meshbrowser/browser`: session state, history, bookmarks, MIME/text extraction, and command dispatch.
- `meshbrowser/agent`: foundational multi-agent coordinator, capabilities, and scoring contracts shared by all navigation transports.
- `meshbrowser/netstack`: standard HTTP(S), Psiphon-tunneled dialing, timeout policy, SCION path evaluation hooks, and connection pooling.
- `meshbrowser/nostr`: Nostr URI parsing, relay pool integration, event fetch/publish, and NIP-aware rendering.
- `meshbrowser/bitchat`: BitChat command parsing, local mesh adapter, kind `20000` fallback publishing, and channel state.
- `meshbrowser/georelays`: relay CSV/JSON loading, scoring, health checks, and user location privacy controls.
- `meshbrowser/scion`: adapter boundary for SCION control-plane queries, path scoring, and path-aware dialing once the upstream SDK is selected.

## Security and privacy notes

- Do not import private keys from upstream demos by default; require explicit user-provided keys or ephemeral identities.
- Local mesh discovery must be opt-in because it can reveal physical proximity.
- Geo-aware relay selection should support coarse regions and manual relay pinning so users are not forced to disclose precise location.
- HTML rendering should default to sanitized text until a safer terminal renderer is selected.

## Multi-agent foundation

The initial Go implementation adds a deterministic coordinator in `meshbrowser/agent`. Each agent advertises capabilities such as web fetches, Nostr resolution, BitChat mesh operations, or SCION route planning. The terminal command asks the coordinator for a plan before rendering, which keeps future upstream integrations additive instead of requiring the command entry point to know every transport detail.

SCION should be treated as a routing specialist rather than a replacement for content protocols: it can score authenticated paths, propose multipath egress, and then hand the selected path to the HTTP, Nostr, or BitChat agent that owns protocol semantics.
