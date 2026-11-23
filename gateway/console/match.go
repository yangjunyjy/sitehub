package console

import (
	"encoding/json"
	"io"
	"net/http"

	"sitehub/gateway/diagnostics"
)

func (c *console) match(w http.ResponseWriter, r *http.Request, _ account) {
	var input diagnostics.MatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		problem(w, 400, "请填写主机、方法、路径和协议")
		return
	}
	result, status, err := c.client.Match(r.Context(), input)
	if err != nil {
		problem(w, status, err.Error())
		return
	}
	reply(w, 200, result)
}
