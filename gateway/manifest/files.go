package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sitehub/gateway/plugin"

	"gopkg.in/yaml.v3"
)

const (
	maxFileBytes = 4 << 20
	maxFiles     = 128
)

// fileDocument is the authoring format. Includes and implicit defaults are
// resolved before a Document is validated, published, or compared by revision.
type fileDocument struct {
	Version   *int                       `json:"version"`
	Include   []string                   `json:"include"`
	Plugins   Plugins                    `json:"plugins"`
	Upstreams map[string]json.RawMessage `json:"upstreams"`
	Profiles  map[string]Plugins         `json:"profiles"`
	Routes    []Route                    `json:"routes"`
	Streams   []Stream                   `json:"streams"`
}

type fileLoader struct {
	document Document
	seen     map[string]bool
	active   map[string]bool
	origins  map[string]string
	bytes    int64
}

// Read loads one JSON or YAML file and its includes using the built-in plugins.
func Read(file string) (Document, error) {
	return ReadWithRegistry(file, plugin.DefaultRegistry())
}

// ReadWithRegistry uses the same plugin registry as the gateway request engine.
// Includes are relative to their containing file; definitions cannot silently
// overwrite one another. File defaults never relax Decode's wire validation.
func ReadWithRegistry(file string, registry *plugin.Registry) (Document, error) {
	if registry == nil {
		return Document{}, fmt.Errorf("plugin registry is required")
	}
	loader := fileLoader{
		document: Document{Version: 1, Upstreams: map[string]Upstream{}, Profiles: map[string]Plugins{}, Plugins: Plugins{}},
		seen:     map[string]bool{}, active: map[string]bool{}, origins: map[string]string{},
	}
	if err := loader.load(file, true); err != nil {
		return Document{}, err
	}
	if len(Bytes(loader.document)) > maxFileBytes {
		return Document{}, fmt.Errorf("resolved manifest exceeds 4 MiB")
	}
	if err := Validate(loader.document, registry); err != nil {
		return Document{}, fmt.Errorf("gateway manifest %s: %w", file, err)
	}
	return loader.document, nil
}

func (l *fileLoader) load(file string, root bool) error {
	absolute, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return fmt.Errorf("gateway configuration %s: %w", file, err)
	}
	// Windows file names are case-insensitive; use one identity for cycle checks.
	identity := filepath.Clean(absolute)
	if filepath.Separator == '\\' {
		identity = strings.ToLower(identity)
	}
	if l.active[identity] {
		return fmt.Errorf("include cycle at %s", absolute)
	}
	if l.seen[identity] {
		return fmt.Errorf("configuration file included more than once: %s", absolute)
	}
	if len(l.seen) >= maxFiles {
		return fmt.Errorf("gateway configuration includes exceed %d files", maxFiles)
	}
	l.seen[identity], l.active[identity] = true, true
	defer delete(l.active, identity)
	input, err := os.Open(absolute)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(io.LimitReader(input, maxFileBytes-l.bytes+1))
	closeErr := input.Close()
	if err != nil {
		return fmt.Errorf("gateway configuration %s: %w", absolute, err)
	}
	if closeErr != nil {
		return closeErr
	}
	l.bytes += int64(len(data))
	if l.bytes > maxFileBytes {
		return fmt.Errorf("gateway configuration files exceed 4 MiB")
	}
	var part fileDocument
	if err := decodeFile(absolute, data, &part); err != nil {
		return fmt.Errorf("gateway configuration %s: %w", absolute, err)
	}
	if root && part.Version == nil || part.Version != nil && *part.Version != 1 {
		return fmt.Errorf("gateway configuration %s: root version must be 1; included versions, when specified, must be 1", absolute)
	}
	if err := l.merge(part, absolute); err != nil {
		return err
	}
	for _, pattern := range part.Include {
		if pattern == "" {
			return fmt.Errorf("gateway configuration %s: include pattern must not be empty", absolute)
		}
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(filepath.Dir(absolute), pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("gateway configuration %s: invalid include %q: %w", absolute, pattern, err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("gateway configuration %s: include %q matched no files", absolute, pattern)
		}
		sort.Strings(matches)
		for _, included := range matches {
			if err := l.load(included, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *fileLoader) definition(kind, name, file string) error {
	key := kind + "/" + name
	if previous, exists := l.origins[key]; exists {
		return fmt.Errorf("duplicate %s %q in %s (already defined in %s)", kind, name, file, previous)
	}
	l.origins[key] = file
	return nil
}

func (l *fileLoader) merge(part fileDocument, file string) error {
	for name, raw := range part.Upstreams {
		if err := l.definition("upstream", name, file); err != nil {
			return err
		}
		upstream, err := decodeUpstream(raw)
		if err != nil {
			return fmt.Errorf("upstream %q in %s: %w", name, file, err)
		}
		l.document.Upstreams[name] = upstream
	}
	for name, values := range part.Profiles {
		if err := l.definition("profile", name, file); err != nil {
			return err
		}
		l.document.Profiles[name] = values
	}
	for name, value := range part.Plugins {
		if err := l.definition("global plugin", name, file); err != nil {
			return err
		}
		l.document.Plugins[name] = value
	}
	for _, route := range part.Routes {
		if err := l.definition("route", route.ID, file); err != nil {
			return err
		}
		l.document.Routes = append(l.document.Routes, route)
	}
	for _, stream := range part.Streams {
		if err := l.definition("stream", stream.ID, file); err != nil {
			return err
		}
		l.document.Streams = append(l.document.Streams, stream)
	}
	return nil
}

func decodeUpstream(raw json.RawMessage) (Upstream, error) {
	var upstream Upstream
	if err := decodeJSON(raw, &upstream); err != nil {
		return upstream, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return upstream, fmt.Errorf("upstream must be an object")
	}
	defaults := []struct {
		name     string
		value    *int
		fallback int
	}{
		{"connect_timeout_ms", &upstream.ConnectTimeoutMS, 3000},
		{"response_header_timeout_ms", &upstream.ResponseHeaderTimeoutMS, 10000},
		{"io_timeout_ms", &upstream.IOTimeoutMS, 30000},
		{"health_interval_ms", &upstream.HealthIntervalMS, 2000},
		{"health_timeout_ms", &upstream.HealthTimeoutMS, 1000},
		{"health_failures", &upstream.HealthFailures, 2},
	}
	for _, field := range defaults {
		if _, exists := fields[field.name]; !exists {
			*field.value = field.fallback
		}
	}
	if _, exists := fields["health_path"]; !exists {
		upstream.HealthPath = "/healthz"
	}
	if len(upstream.Nodes) > 0 {
		var nodes []map[string]json.RawMessage
		if err := json.Unmarshal(fields["nodes"], &nodes); err != nil {
			return upstream, err
		}
		for index, node := range nodes {
			if _, exists := node["weight"]; !exists {
				upstream.Nodes[index].Weight = 1
			}
		}
	}
	return upstream, nil
}

func decodeFile(file string, data []byte, target any) error {
	if extension := strings.ToLower(filepath.Ext(file)); extension == ".yaml" || extension == ".yml" {
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		var root yaml.Node
		if err := decoder.Decode(&root); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		if decoder.Decode(new(yaml.Node)) != io.EOF {
			return fmt.Errorf("configuration must contain one YAML document")
		}
		if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
			return fmt.Errorf("configuration must be an object")
		}
		value, err := yamlValue(root.Content[0])
		if err != nil {
			return err
		}
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("YAML values must be JSON-compatible: %w", err)
		}
	} else {
		if err := checkJSONKeys(data); err != nil {
			return err
		}
	}
	return decodeJSON(data, target)
}

func decodeJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("configuration must contain one document")
	}
	return nil
}

// Converting the YAML node tree ourselves preserves RawMessage plugin values,
// rejects non-string keys and duplicate definitions, and avoids alias expansion.
func yamlValue(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.MappingNode:
		values := make(map[string]any, len(node.Content)/2)
		for index := 0; index < len(node.Content); index += 2 {
			key := node.Content[index]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return nil, fmt.Errorf("YAML line %d: object keys must be strings; merge keys are not supported", key.Line)
			}
			if _, duplicate := values[key.Value]; duplicate {
				return nil, fmt.Errorf("YAML line %d: duplicate key %q", key.Line, key.Value)
			}
			value, err := yamlValue(node.Content[index+1])
			if err != nil {
				return nil, err
			}
			values[key.Value] = value
		}
		return values, nil
	case yaml.SequenceNode:
		values := make([]any, len(node.Content))
		for index, child := range node.Content {
			value, err := yamlValue(child)
			if err != nil {
				return nil, err
			}
			values[index] = value
		}
		return values, nil
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str", "!!bool", "!!int", "!!float", "!!null":
			var value any
			if err := node.Decode(&value); err != nil {
				return nil, err
			}
			return value, nil
		}
		return nil, fmt.Errorf("YAML line %d: unsupported scalar type %q", node.Line, node.Tag)
	default:
		return nil, fmt.Errorf("YAML line %d: aliases are not supported; use includes or plugin profiles", node.Line)
	}
}

func checkJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var visit func() error
	visit = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, composite := token.(json.Delim)
		if !composite {
			return nil
		}
		if delimiter == '{' {
			keys := map[string]bool{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("object keys must be strings")
				}
				if keys[name] {
					return fmt.Errorf("duplicate JSON key %q", name)
				}
				keys[name] = true
				if err := visit(); err != nil {
					return err
				}
			}
		} else if delimiter == '[' {
			for decoder.More() {
				if err := visit(); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("unexpected JSON delimiter")
		}
		_, err = decoder.Token()
		return err
	}
	if err := visit(); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("configuration must contain one JSON document")
	}
	return nil
}
