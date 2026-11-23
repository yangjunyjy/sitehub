package control

import (
	"encoding/json"
	"io"
	"net/http"

	"sitehub/gateway/diagnostics"
)

func (c *Controller) match(w http.ResponseWriter, r *http.Request) {
	var input diagnostics.MatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		respond(w, 400, map[string]string{"message": "请求匹配参数无效"})
		return
	}
	c.mu.Lock()
	result, err := c.Engine.InspectMatch(input)
	result.Revision = c.applied
	c.mu.Unlock()
	if err != nil {
		respond(w, 400, map[string]string{"message": err.Error()})
		return
	}
	respond(w, 200, result)
}
