// Package publisher manages complete native gateway configurations using CAS.
package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"sitehub/gateway/diagnostics"
	"sitehub/gateway/manifest"
)

type State = diagnostics.State
type ControlError struct {
	Status  int
	Message string
	Outcome string
}

func (e *ControlError) Error() string {
	return fmt.Sprintf("gateway control returned %d: %s", e.Status, e.Message)
}

type Client struct {
	URL, Key string
	HTTP     *http.Client
}

func (c *Client) api(ctx context.Context, method, path string, revision int64, generation string, value, result any) error {
	var body io.Reader
	if value != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.URL, "/")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	if generation != "" {
		req.Header.Set("X-Release-ID", generation)
	}
	if method == http.MethodPut {
		req.Header.Set("If-Match", strconv.Quote(strconv.FormatInt(revision, 10)))
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("gateway control request failed; query current revision before retrying")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var detail struct{ Message, Outcome string }
		_ = json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&detail)
		if detail.Message == "" {
			detail.Message = http.StatusText(resp.StatusCode)
		}
		return &ControlError{Status: resp.StatusCode, Message: detail.Message, Outcome: detail.Outcome}
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(result); err != nil {
		return fmt.Errorf("invalid gateway control response")
	}
	return nil
}
func (c *Client) request(ctx context.Context, method string, revision int64, generation string, value any) (State, error) {
	var state State
	if err := c.api(ctx, method, "/v1/config", revision, generation, value, &state); err != nil {
		return state, err
	}
	if state.Revision < 1 {
		return state, fmt.Errorf("gateway returned invalid revision")
	}
	return state, nil
}
func (c *Client) Snapshot(ctx context.Context) (State, error) {
	return c.request(ctx, http.MethodGet, 0, "", nil)
}
func (c *Client) Workspace(ctx context.Context) (State, error) {
	var state State
	err := c.api(ctx, http.MethodGet, "/v1/config?local_fallback=1", 0, "", nil, &state)
	if err == nil && state.Revision < 1 && !(state.DesiredUnavailable && state.AppliedManifest != nil && state.AppliedRevision > 0) {
		err = fmt.Errorf("gateway returned invalid workspace revision")
	}
	return state, err
}
func (c *Client) Publish(ctx context.Context, desired manifest.Document, expected int64) (State, error) {
	return c.request(ctx, http.MethodPut, expected, "", desired)
}
func Diff(desired, previous manifest.Document) []string {
	flatten := func(d manifest.Document) map[string]string {
		out := map[string]string{}
		if len(d.Plugins) > 0 {
			b, _ := json.Marshal(d.Plugins)
			out["plugins/global"] = string(b)
		}
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
		for _, stream := range d.Streams {
			b, _ := json.Marshal(stream)
			out["stream/"+stream.ID] = string(b)
		}
		return out
	}
	old, next := flatten(previous), flatten(desired)
	var changes []string
	for k, v := range next {
		if prior, ok := old[k]; !ok {
			changes = append(changes, "+ "+k)
		} else if !equalResource(prior, v) {
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

// Plugin parameters are RawMessage values. Their object keys can arrive in a
// different order from the form editor without changing the configuration.
func equalResource(a, b string) bool {
	if a == b {
		return true
	}
	decode := func(raw string) any {
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		var value any
		_ = decoder.Decode(&value) // Diff receives marshalled manifest resources.
		return value
	}
	return reflect.DeepEqual(decode(a), decode(b))
}
