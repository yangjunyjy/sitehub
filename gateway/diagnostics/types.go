// Package diagnostics describes the local gateway's read-only inspection API.
package diagnostics

import "sitehub/gateway/manifest"

// State pairs the store's desired configuration with this instance's applied
// revision. A read-only workspace can still inspect local traffic during a
// store outage. Revision is then unknown and cannot serve as a CAS base.
type State struct {
	Runtime            *Runtime           `json:"runtime,omitempty"`
	AppliedManifest    *manifest.Document `json:"applied_manifest,omitempty"`
	DesiredUnavailable bool               `json:"desired_unavailable,omitempty"`
	FileManaged        bool               `json:"file_managed,omitempty"`
	Revision           int64              `json:"revision"`
	AppliedRevision    int64              `json:"applied_revision"`
	Manifest           manifest.Document  `json:"manifest"`
	LastError          string             `json:"last_error,omitempty"`
}

type MatchRequest struct {
	Host   string `json:"host"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Scheme string `json:"scheme"`
}

type MatchResult struct {
	Revision int64  `json:"revision"`
	Matched  bool   `json:"matched"`
	Reason   string `json:"reason"`
	RouteID  string `json:"route_id,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Upstream string `json:"upstream,omitempty"`
	Profile  string `json:"profile,omitempty"`
	Priority int    `json:"priority"`
	Terminal string `json:"terminal,omitempty"`
}

type Node struct {
	URL     string `json:"url"`
	Healthy bool   `json:"healthy"`
	Weight  int    `json:"weight"`
}

type Listener struct {
	manifest.Stream
	Address string `json:"address"` // kernel-reported local address
}

type Runtime struct {
	PathWarnings []string          `json:"path_warnings,omitempty"`
	HTTPListen   string            `json:"http_listen"`
	Upstreams    map[string][]Node `json:"upstreams"`
	Streams      []Listener        `json:"streams"`
}
