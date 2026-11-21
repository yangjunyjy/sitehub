package logic

import (
	"encoding/json"
	"sitehub/service/blog/api/internal/types"
)

func Reply(value any, err error) (*types.Response, error) {
	if err != nil {
		return nil, Classify(err)
	}
	return &types.Response{Message: "ok", Data: articleResponse(value)}, nil
}
func Done(err error) (*types.Response, error) { return Reply(struct{}{}, err) }
func TextOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func DecodeRequest(source, target any) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func ReplyAs[T any](value any, err error) (*types.Response, error) {
	if err != nil {
		return nil, Classify(err)
	}
	var payload T
	if err := DecodeRequest(value, &payload); err != nil {
		return nil, err
	}
	return Reply(payload, nil)
}
