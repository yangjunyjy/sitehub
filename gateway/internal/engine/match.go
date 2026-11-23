package engine

import (
	"errors"
	"net/url"
	"path"
	"slices"
	"strings"

	"golang.org/x/net/http/httpguts"
	"sitehub/gateway/diagnostics"
	"sitehub/gateway/manifest"
)

// Both live traffic and console inspection use the same path guards and
// winner selection. Inspection never invokes plugins or contacts an upstream.
func (s *Snapshot) match(authority, method, scheme string, u *url.URL) (*route, string, string, error) {
	p := u.Path
	raw := strings.ToLower(u.EscapedPath())
	if strings.ContainsAny(p, "\\\x00\r\n") || strings.Contains(raw, "%2f") || strings.Contains(raw, "%5c") || strings.Contains(raw, "%25") || strings.Contains(p, "//") || (path.Clean(p) != p && path.Clean(p)+"/" != p) {
		return nil, "", "", errors.New("Invalid request path")
	}
	host, err := manifest.ParseHost(authority)
	if err != nil {
		return nil, "", "", errors.New("Invalid request host")
	}
	if host.Port == "" {
		host.Port = "80"
		if scheme == "https" {
			host.Port = "443"
		}
	}

	var winner *route
	var winningPattern string
	bestLength := -1
	bestExact := false
	hostMatched, methodMatched := false, false
	for i := range s.routes {
		candidate := &s.routes[i]
		c := candidate.config
		// Prepare sorts by priority, host specificity and ID. A later
		// priority/specificity group cannot displace an existing match.
		if winner != nil && (c.Priority != winner.config.Priority || candidate.host.Specificity() != winner.host.Specificity()) {
			break
		}
		if !candidate.host.Matches(host) {
			continue
		}
		hostMatched = true
		if len(c.Methods) > 0 && !slices.Contains(c.Methods, method) {
			continue
		}
		methodMatched = true
		for _, pattern := range c.Paths {
			literal, prefix := strings.CutSuffix(pattern, "*")
			exact := !prefix
			if (exact && p != literal) || (prefix && !strings.HasPrefix(p, literal)) {
				continue
			}
			if len(literal) > bestLength || (len(literal) == bestLength && exact && !bestExact) {
				winner, winningPattern = candidate, pattern
				bestLength, bestExact = len(literal), exact
			}
		}
	}
	switch {
	case winner != nil:
		return winner, winningPattern, "", nil
	case methodMatched:
		return nil, "", "主机和方法已匹配，但路径没有命中", nil
	case hostMatched:
		return nil, "", "主机已匹配，但没有规则接受此请求方法", nil
	default:
		return nil, "", "没有规则匹配此 Host 和端口", nil
	}
}

func (e *Engine) InspectMatch(input diagnostics.MatchRequest) (diagnostics.MatchResult, error) {
	var result diagnostics.MatchResult
	if (input.Scheme != "http" && input.Scheme != "https") || input.Host == "" || strings.Contains(input.Host, "*") || !httpguts.ValidHeaderFieldName(input.Method) {
		return result, errors.New("请填写具体的主机、HTTP 方法和 http 或 https 协议")
	}
	u, err := url.ParseRequestURI(input.Path)
	if err != nil || !strings.HasPrefix(input.Path, "/") || strings.Contains(input.Path, "#") {
		return result, errors.New("请填写以 / 开头的请求路径")
	}
	s := e.current.Load()
	if s == nil {
		return result, errors.New("网关尚未应用配置")
	}
	winner, pattern, reason, err := s.match(input.Host, input.Method, input.Scheme, u)
	if err != nil || winner == nil {
		result.Reason = reason
		return result, err
	}
	r := winner.config
	result = diagnostics.MatchResult{Matched: true, Reason: "已命中规则", RouteID: r.ID, Pattern: pattern, Profile: r.Profile, Upstream: r.Upstream, Priority: r.Priority}
	if r.Upstream == "" {
		result.Terminal, _ = e.registry.TerminalName(manifest.EffectivePlugins(s.Document, r))
	}
	return result, nil
}
