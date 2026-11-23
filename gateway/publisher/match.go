package publisher

import (
	"context"
	"errors"
	"net/http"

	"sitehub/gateway/diagnostics"
)

// Match asks the gateway owning the live snapshot, including when the console
// runs as a standalone gatewayctl process.
func (c *Client) Match(ctx context.Context, input diagnostics.MatchRequest) (diagnostics.MatchResult, int, error) {
	var result diagnostics.MatchResult
	err := c.api(ctx, http.MethodPost, "/v1/match", 0, "", input, &result)
	var detail *ControlError
	if errors.As(err, &detail) && detail.Status == http.StatusBadRequest {
		return result, detail.Status, errors.New(detail.Message)
	}
	if err != nil || result.Revision < 1 {
		return result, http.StatusBadGateway, errors.New("无法读取生效规则的匹配结果")
	}
	return result, http.StatusOK, nil
}
