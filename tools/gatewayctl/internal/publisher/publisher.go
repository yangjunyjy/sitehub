// Package publisher manages complete native gateway configurations using CAS.
package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"sitehub/gateway/manifest"
)

type State struct {
	Revision        int64             `json:"revision"`
	AppliedRevision int64             `json:"applied_revision"`
	Manifest        manifest.Document `json:"manifest"`
	LastError       string            `json:"last_error,omitempty"`
}
type Client struct {
	URL, Key string
	HTTP     *http.Client
}

func (c *Client) request(ctx context.Context, method string, revision int64, value any) (State, error) {
	var state State
	var body io.Reader
	if value != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return state, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.URL, "/")+"/v1/config", body)
	if err != nil {
		return state, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	if method == http.MethodPut {
		req.Header.Set("If-Match", strconv.Quote(strconv.FormatInt(revision, 10)))
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return state, fmt.Errorf("gateway control request failed; query current revision before retrying")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return state, fmt.Errorf("gateway control returned %d; query active revision before retrying", resp.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 8<<20))
	if err = decoder.Decode(&state); err != nil {
		return state, fmt.Errorf("invalid gateway control response")
	}
	if state.Revision < 1 {
		return state, fmt.Errorf("gateway returned invalid revision")
	}
	return state, nil
}
func (c *Client) Snapshot(ctx context.Context) (State, error) {
	return c.request(ctx, http.MethodGet, 0, nil)
}
func (c *Client) Publish(ctx context.Context, desired manifest.Document, expected int64) (State, error) {
	return c.request(ctx, http.MethodPut, expected, desired)
}
func Diff(desired, previous manifest.Document) []string {
	flatten := func(d manifest.Document) map[string]string {
		out := map[string]string{}
		for k, v := range d.Upstreams {
			b, _ := json.Marshal(v)
			out["upstream/"+k] = string(b)
		}
		for k, v := range d.Profiles {
			b, _ := json.Marshal(v)
			out["profile/"+k] = string(b)
		}
		for _, v := range d.Routes {
			b, _ := json.Marshal(v)
			out["route/"+v.ID] = string(b)
		}
		return out
	}
	old, next := flatten(previous), flatten(desired)
	var changes []string
	for k, v := range next {
		if prior, ok := old[k]; !ok {
			changes = append(changes, "+ "+k)
		} else if prior != v {
			changes = append(changes, "~ "+k)
		}
		delete(old, k)
	}
	for k := range old {
		changes = append(changes, "- "+k)
	}
	sort.Strings(changes)
	return changes
}
