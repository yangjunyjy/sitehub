package manifest

import (
	"fmt"
	"slices"
	"strings"
)

// PathWarnings identifies overlapping rules affected by exact-before-prefix
// precedence. It also works offline, before a candidate claims any sockets.
func PathWarnings(routes []Route) []string {
	type scope struct {
		host     HostMatcher
		priority int
		path     string
	}
	prefixes := make(map[scope][]Route)
	for _, r := range routes {
		host, _ := ParseHost(r.Host) // The input is a validated manifest.
		for _, pattern := range r.Paths {
			if base, ok := strings.CutSuffix(pattern, "*"); ok {
				key := scope{host, r.Priority, base}
				prefixes[key] = append(prefixes[key], r)
			}
		}
	}
	var warnings []string
	for _, r := range routes {
		host, _ := ParseHost(r.Host)
		for _, pattern := range r.Paths {
			for _, other := range prefixes[scope{host, r.Priority, pattern}] {
				if r.ID == other.ID || !methodsOverlap(r.Methods, other.Methods) {
					continue
				}
				if len(warnings) == 20 {
					return append(warnings, "还有其他重叠规则，此处仅列出前 20 组。")
				}
				warnings = append(warnings, fmt.Sprintf("路径 %s 同时符合 %s 的精确规则与 %s 的前缀规则，精确规则优先。", pattern, r.ID, other.ID))
			}
		}
	}
	return warnings
}

func methodsOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return true
	}
	for _, method := range a {
		if slices.Contains(b, method) {
			return true
		}
	}
	return false
}
