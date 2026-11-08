// Package manifest defines the native gateway configuration contract. A document
// is validated as a whole before it can replace the running request pipeline.
package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"sitehub/gateway/plugin"
)

type Plugins map[string]json.RawMessage
type Document struct {
	Version   int                 `json:"version"`
	Upstreams map[string]Upstream `json:"upstreams"`
	Profiles  map[string]Plugins  `json:"profiles"`
	Routes    []Route             `json:"routes"`
}
type Node struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}
type Upstream struct {
	Service                 string `json:"service,omitempty"`
	Nodes                   []Node `json:"nodes,omitempty"`
	ConnectTimeoutMS        int    `json:"connect_timeout_ms"`
	ResponseHeaderTimeoutMS int    `json:"response_header_timeout_ms"`
	IOTimeoutMS             int    `json:"io_timeout_ms"`
	HealthPath              string `json:"health_path"`
	HealthIntervalMS        int    `json:"health_interval_ms"`
	HealthTimeoutMS         int    `json:"health_timeout_ms"`
	HealthFailures          int    `json:"health_failures"`
}
type Route struct {
	ID       string   `json:"id"`
	Host     string   `json:"host"`
	Paths    []string `json:"paths"`
	Methods  []string `json:"methods,omitempty"`
	Priority int      `json:"priority"`
	Upstream string   `json:"upstream,omitempty"`
	Profile  string   `json:"profile,omitempty"`
	Plugins  Plugins  `json:"plugins,omitempty"`
}

var identifier = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
var serviceName = regexp.MustCompile(`^[a-z][a-z0-9-]*-api$`)

func Decode(data []byte) (Document, error) {
	var d Document
	if len(data) > 4<<20 {
		return d, fmt.Errorf("manifest exceeds 4 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&d); err != nil {
		return d, fmt.Errorf("invalid gateway manifest: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return d, fmt.Errorf("manifest must contain one JSON document")
	}
	return d, nil
}
func Read(file string) (Document, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Document{}, err
	}
	d, err := Decode(data)
	if err != nil {
		return d, err
	}
	return d, Validate(d, plugin.DefaultRegistry())
}
func Bytes(d Document) []byte    { b, _ := json.Marshal(d); return b }
func Revision(d Document) string { sum := sha256.Sum256(Bytes(d)); return hex.EncodeToString(sum[:]) }
func EffectivePlugins(d Document, route Route) Plugins {
	result := Plugins{}
	for name, config := range d.Profiles[route.Profile] {
		result[name] = config
	}
	for name, config := range route.Plugins {
		result[name] = config
	}
	return result
}
func Validate(d Document, plugins *plugin.Registry) error {
	if d.Version != 1 || len(d.Routes) == 0 || len(d.Routes) > 1000 || len(d.Upstreams) > 256 || len(d.Profiles) > 256 {
		return fmt.Errorf("manifest version must be 1 with 1..1000 routes and at most 256 upstreams/profiles")
	}
	for id, u := range d.Upstreams {
		if !identifier.MatchString(id) || (u.Service == "") == (len(u.Nodes) == 0) {
			return fmt.Errorf("upstream %s needs exactly one of service or nodes", id)
		}
		if u.Service != "" && !serviceName.MatchString(u.Service) {
			return fmt.Errorf("upstream %s must discover an API service", id)
		}
		if len(u.Nodes) > 128 {
			return fmt.Errorf("too many nodes in upstream %s", id)
		}
		seen := map[string]bool{}
		for _, node := range u.Nodes {
			v, err := url.Parse(node.URL)
			if err != nil || (v.Scheme != "http" && v.Scheme != "https") || v.Hostname() == "" || v.User != nil || v.RawQuery != "" || v.Fragment != "" || (v.Path != "" && v.Path != "/") || node.Weight < 1 || node.Weight > 1000 || seen[node.URL] {
				return fmt.Errorf("invalid node in upstream %s", id)
			}
			seen[node.URL] = true
		}
		if u.IOTimeoutMS < 1 || u.IOTimeoutMS > 3600000 || u.ConnectTimeoutMS < 1 || u.ConnectTimeoutMS > 60000 || u.ResponseHeaderTimeoutMS < 1 || u.ResponseHeaderTimeoutMS > 3600000 || u.HealthIntervalMS < 100 || u.HealthIntervalMS > 60000 || u.HealthTimeoutMS < 1 || u.HealthTimeoutMS > 60000 || u.HealthFailures < 1 || u.HealthFailures > 100 || !validPath(u.HealthPath, false) {
			return fmt.Errorf("invalid timeout or health configuration for %s", id)
		}
	}
	for name, values := range d.Profiles {
		if !identifier.MatchString(name) {
			return fmt.Errorf("invalid profile id")
		}
		if err := plugins.Validate(values); err != nil {
			return fmt.Errorf("profile %s: %w", name, err)
		}
	}
	seen := map[string]bool{}
	for _, r := range d.Routes {
		if !identifier.MatchString(r.ID) || seen[r.ID] {
			return fmt.Errorf("invalid or duplicate route id %s", r.ID)
		}
		seen[r.ID] = true
		if r.Host == "" || r.Host != strings.ToLower(r.Host) || strings.ContainsAny(r.Host, "* /:\\\t\r\n") || net.ParseIP(r.Host) != nil {
			return fmt.Errorf("route %s needs a lowercase explicit DNS host without port", r.ID)
		}
		if len(r.Paths) == 0 || len(r.Paths) > 64 {
			return fmt.Errorf("route %s needs paths", r.ID)
		}
		for _, p := range r.Paths {
			if !validPath(p, true) {
				return fmt.Errorf("invalid route path in %s", r.ID)
			}
		}
		for _, method := range r.Methods {
			if method != "GET" && method != "HEAD" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" && method != "OPTIONS" {
				return fmt.Errorf("unsupported method in %s", r.ID)
			}
		}
		if r.Profile != "" {
			if _, ok := d.Profiles[r.Profile]; !ok {
				return fmt.Errorf("unknown profile for %s", r.ID)
			}
		}
		effective := EffectivePlugins(d, r)
		if err := plugins.Validate(effective); err != nil {
			return fmt.Errorf("route %s: %w", r.ID, err)
		}
		_, terminal := effective["static-response"]
		if terminal == (r.Upstream != "") {
			return fmt.Errorf("route %s needs exactly one upstream or static-response", r.ID)
		}
		if r.Upstream != "" {
			if _, ok := d.Upstreams[r.Upstream]; !ok {
				return fmt.Errorf("unknown upstream for %s", r.ID)
			}
		}
	}
	return nil
}
func validPath(value string, wildcard bool) bool {
	if value == "" || !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#%\\\r\n") {
		return false
	}
	if wildcard && strings.HasSuffix(value, "/*") {
		value = strings.TrimSuffix(value, "*")
	}
	if strings.Contains(value, "*") || strings.Contains(value, "//") {
		return false
	}
	return path.Clean(value) == value || (strings.HasSuffix(value, "/") && path.Clean(value)+"/" == value)
}
