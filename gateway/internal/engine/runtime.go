package engine

import (
	"sort"
	"strings"

	"sitehub/gateway/diagnostics"
	"sitehub/gateway/manifest"
)

// Runtime reports only this process. Listener addresses are sampled from bound
// sockets and node health from the same pools used for forwarding.
func (e *Engine) Runtime(httpAddress string) (*diagnostics.Runtime, *manifest.Document) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.current.Load()
	if s == nil {
		return nil, nil
	}
	r := &diagnostics.Runtime{HTTPListen: httpAddress, Upstreams: make(map[string][]diagnostics.Node, len(s.pools)), Streams: []diagnostics.Listener{}}
	r.PathWarnings = s.pathWarnings
	for name, pool := range s.pools {
		pool.mu.Lock()
		nodes := make([]diagnostics.Node, 0, len(pool.nodes))
		for _, node := range pool.nodes {
			nodes = append(nodes, diagnostics.Node{URL: node.target.String(), Healthy: node.healthy, Weight: node.weight})
		}
		pool.mu.Unlock()
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].URL < nodes[j].URL })
		r.Upstreams[name] = nodes
	}
	for _, rule := range s.Document.Streams {
		socket := e.streamSockets[rule.Protocol+"/"+strings.ToLower(rule.Listen)]
		listener := diagnostics.Listener{Stream: rule}
		if socket != nil {
			if socket.listener != nil {
				listener.Address = socket.listener.Addr().String()
			} else {
				listener.Address = socket.packet.LocalAddr().String()
			}
			r.Streams = append(r.Streams, listener)
		}
	}
	return r, &s.Document
}
