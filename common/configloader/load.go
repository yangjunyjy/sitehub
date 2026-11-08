// Package configloader merges service configuration without constructing runtime dependencies.
package configloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/conf"
)

// Load uses Viper for configuration sources and precedence. go-zero only decodes
// the merged values, preserving its embedded fields, defaults and validation tags.
// Each call must receive a fresh Viper and a pointer to a new configuration struct.
func Load(settings *viper.Viper, envPrefix string, target any) error {
	return LoadRemote(settings, envPrefix, target, nil)
}

// LoadRemote adds a remote YAML layer below explicit flags and environment.
// A fresh Viper and target are required for every candidate, including reloads.
func LoadRemote(settings *viper.Viper, envPrefix string, target any, remote []byte) error {
	typ := reflect.TypeOf(target)
	if typ == nil || typ.Kind() != reflect.Pointer || reflect.ValueOf(target).IsNil() || typ.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configuration target must be a non-nil pointer to a struct")
	}
	if envPrefix == "" {
		return fmt.Errorf("configuration environment prefix is required")
	}
	if err := settings.ReadInConfig(); err != nil {
		return fmt.Errorf("read configuration: %w", err)
	}
	// Expand parsed file values, not YAML text: quotes, colons and newlines in
	// credentials remain literal data. Environment overrides are not expanded again.
	fileValues := settings.AllSettings()
	expandFileValues(fileValues)
	if err := settings.MergeConfigMap(fileValues); err != nil {
		return fmt.Errorf("expand configuration: %w", err)
	}
	if len(remote) > 0 {
		layer := viper.New()
		layer.SetConfigType("yaml")
		if err := layer.ReadConfig(bytes.NewReader(remote)); err != nil {
			return fmt.Errorf("invalid remote YAML")
		}
		if len(layer.AllSettings()) == 0 {
			return fmt.Errorf("remote configuration must be a non-empty mapping")
		}
		known := make(map[string]reflect.Type)
		configFields(typ.Elem(), "", known)
		if err := validateRemoteKeys(layer.AllSettings(), "", known); err != nil {
			return err
		}
		// Connection credentials and the config source are local bootstrap data.
		for _, key := range []string{"nacos", "configcenter"} {
			if layer.IsSet(key) {
				return fmt.Errorf("remote configuration cannot set %s", key)
			}
		}
		if err := settings.MergeConfigMap(layer.AllSettings()); err != nil {
			return fmt.Errorf("merge remote configuration: %w", err)
		}
	}
	settings.SetEnvPrefix(envPrefix)
	settings.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	settings.AllowEmptyEnv(true)
	fields := make(map[string]reflect.Type)
	configFields(typ.Elem(), "", fields)
	for key := range fields {
		// Explicit binding includes fields absent from the file. AutomaticEnv alone
		// would omit these fields from AllSettings, silently losing overrides.
		if err := settings.BindEnv(key); err != nil {
			return fmt.Errorf("bind configuration field %s: %w", key, err)
		}
	}
	values := settings.AllSettings()
	for key, fieldType := range fields {
		if err := coerceScalar(values, key, fieldType); err != nil {
			return err
		}
	}
	content, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("encode configuration: %w", err)
	}
	if err := conf.LoadFromJsonBytes(content, target); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}
	return nil
}

func validateRemoteKeys(values map[string]any, prefix string, known map[string]reflect.Type) error {
	for name, value := range values {
		key := prefix + name
		if _, ok := known[key]; ok {
			continue
		}
		found := false
		for field := range known {
			if strings.HasPrefix(field, key+".") {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown remote configuration field %s", key)
		}
		if nested, ok := value.(map[string]any); ok {
			if err := validateRemoteKeys(nested, key+".", known); err != nil {
				return err
			}
		}
	}
	return nil
}

func expandFileValues(values map[string]any) {
	for key, value := range values {
		values[key] = expandFileValue(value)
	}
}

func expandFileValue(value any) any {
	switch value := value.(type) {
	case string:
		return os.ExpandEnv(value)
	case map[string]any:
		expandFileValues(value)
	case []any:
		for i := range value {
			value[i] = expandFileValue(value[i])
		}
	}
	return value
}

// configFields follows the JSON field names used by go-zero. Anonymous framework
// configs are flattened; nested structs use dot-separated Viper keys.
func configFields(typ reflect.Type, prefix string, fields map[string]reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			continue
		}
		fieldType := field.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if field.Anonymous && name == "" && fieldType.Kind() == reflect.Struct {
			configFields(fieldType, prefix, fields)
			continue
		}
		if name == "" {
			name = field.Name
		}
		key := prefix + strings.ToLower(name)
		if fieldType.Kind() == reflect.Struct {
			configFields(fieldType, key+".", fields)
		} else {
			fields[key] = fieldType
		}
	}
}

// Viper environment values are strings. Convert numeric and boolean fields
// before go-zero's strict decoder; duration and JSON collection strings are
// handled by go-zero itself. Never include a configuration value in errors here.
func coerceScalar(values map[string]any, key string, typ reflect.Type) error {
	parts := strings.Split(key, ".")
	for _, part := range parts[:len(parts)-1] {
		nested, ok := values[part].(map[string]any)
		if !ok {
			return nil
		}
		values = nested
	}
	name := parts[len(parts)-1]
	value, ok := values[name].(string)
	if !ok || typ == reflect.TypeOf(time.Duration(0)) {
		return nil
	}
	switch typ.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("configuration field %s must be a boolean", key)
		}
		values[name] = parsed
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		var parsed json.Number
		if err := json.Unmarshal([]byte(value), &parsed); err != nil || parsed == "" {
			return fmt.Errorf("configuration field %s must be a number", key)
		}
		values[name] = parsed
	}
	return nil
}
