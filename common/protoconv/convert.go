// Package protoconv maps typed domain values to protobuf messages. It preserves
// integer precision and optional scalar presence across the JSON representation.
package protoconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func key(s string) string { return strings.ToLower(strings.ReplaceAll(s, "_", "")) }

func readJSON(data []byte) (any, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err := decoder.Decode(&value)
	return value, err
}

// Encode projects a domain value onto an explicit protobuf response schema.
// Fields absent from that schema (for example password hashes) are never emitted.
func Encode(value any, target proto.Message) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	v, err := readJSON(data)
	if err != nil {
		return err
	}
	v = protobufKeys(v, target.ProtoReflect().Descriptor())
	data, err = json.Marshal(v)
	if err != nil {
		return err
	}
	return (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(data, target)
}

func protobufKeys(value any, descriptor protoreflect.MessageDescriptor) any {
	if strings.HasPrefix(string(descriptor.FullName()), "google.protobuf.") {
		return value
	}
	object, ok := value.(map[string]any)
	if !ok {
		return value
	}
	result := make(map[string]any)
	fields := descriptor.Fields()
	for name, v := range object {
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if key(name) != key(field.JSONName()) && key(name) != key(string(field.Name())) {
				continue
			}
			if field.Message() != nil && !field.IsMap() {
				if field.IsList() {
					if items, ok := v.([]any); ok {
						for j := range items {
							items[j] = protobufKeys(items[j], field.Message())
						}
					}
				} else {
					v = protobufKeys(v, field.Message())
				}
			}
			result[field.JSONName()] = v
			break
		}
	}
	return result
}

// Decode converts protobuf integer strings to JSON numbers only when the target
// field is numeric. String IDs and timestamps retain their original meaning.
func Decode(source proto.Message, target any) error {
	t := reflect.TypeOf(target)
	if t == nil || t.Kind() != reflect.Pointer || reflect.ValueOf(target).IsNil() {
		return fmt.Errorf("decode target must be a non-nil pointer")
	}
	data, err := protojson.Marshal(source)
	if err != nil {
		return err
	}
	v, err := readJSON(data)
	if err != nil {
		return err
	}
	v = goKeys(v, t.Elem())
	data, err = json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func goKeys(value any, t reflect.Type) any {
	if value == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		return value
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		if s, ok := value.(string); ok {
			return json.Number(s)
		}
	case reflect.Slice, reflect.Array:
		if items, ok := value.([]any); ok {
			for i := range items {
				items[i] = goKeys(items[i], t.Elem())
			}
		}
	case reflect.Struct:
		if object, ok := value.(map[string]any); ok {
			fields := map[string]reflect.StructField{}
			var collect func(reflect.Type)
			collect = func(typ reflect.Type) {
				for i := 0; i < typ.NumField(); i++ {
					f := typ.Field(i)
					if f.PkgPath != "" {
						continue
					}
					name := strings.Split(f.Tag.Get("json"), ",")[0]
					if name == "-" {
						continue
					}
					if f.Anonymous && name == "" {
						nested := f.Type
						for nested.Kind() == reflect.Pointer {
							nested = nested.Elem()
						}
						if nested.Kind() == reflect.Struct {
							collect(nested)
							continue
						}
					}
					if name == "" {
						name = f.Name
					}
					fields[name] = f
				}
			}
			collect(t)
			result := map[string]any{}
			for name, v := range object {
				for jsonName, f := range fields {
					if key(name) == key(jsonName) || key(name) == key(f.Name) {
						result[jsonName] = goKeys(v, f.Type)
						break
					}
				}
			}
			return result
		}
	}
	return value
}

func Value(message proto.Message) (any, error) {
	data, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(message)
	if err != nil {
		return nil, err
	}
	return readJSON(data)
}
