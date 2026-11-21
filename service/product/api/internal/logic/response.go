package logic

import (
	"encoding/json"

	"sitehub/service/product/api/internal/types"
)

func CopyManifest(source, destination any) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, destination)
}

func Response(data any, err error) (*types.Response, error) {
	if err != nil {
		return nil, classify(err)
	}
	return &types.Response{Code: 0, Message: "ok", Data: data}, nil
}
