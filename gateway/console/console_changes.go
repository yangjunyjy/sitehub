package console

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"sitehub/common/rpcx"
	"sitehub/gateway/manifest"
	"sitehub/gateway/publisher"
)

type changeRequest struct {
	Password   string          `json:"password,omitempty"`
	Revision   int64           `json:"revision"`
	Manifest   json.RawMessage `json:"manifest"`
	Generation string          `json:"generation"`
}

func (c *console) change(w http.ResponseWriter, r *http.Request, actor account) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		problem(w, 415, "请提交 JSON")
		return
	}
	var req changeRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, (4<<20)+4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&req) != nil || decoder.Decode(new(any)) != io.EOF || req.Revision < 1 {
		problem(w, 400, "配置或修订号无效")
		return
	}
	if r.URL.Path != "/api/preview" && !c.identity.checkPassword(w, rpcx.HTTPValue(r.Context(), "x-client-ip"), actor, req.Password) {
		return
	}
	state, err := c.client.Snapshot(r.Context())
	if err != nil {
		problem(w, 502, "无法读取网关配置")
		return
	}
	if req.Revision != state.Revision {
		problem(w, 409, "网关配置已更新，请重新载入当前修订后再提交")
		return
	}
	preview := r.URL.Path == "/api/preview"
	if !preview && state.FileManaged {
		problem(w, 409, "网关由文件管理，页面发布需要 etcd 模式")
		return
	}
	var desired manifest.Document
	if r.URL.Path == "/api/rollback" || (preview && req.Generation != "") {
		if !actor.allows(rollbackConfig) {
			problem(w, 403, "当前网关账号没有回滚权限")
			return
		}
		if !publisher.GenerationID.MatchString(req.Generation) || len(req.Manifest) > 0 {
			problem(w, 400, "请选择发布记录")
			return
		}
		saved, err := publisher.LoadRelease(filepath.Join(c.dir, req.Generation+".json"))
		if errors.Is(err, os.ErrNotExist) {
			problem(w, 404, "找不到这次发布")
			return
		}
		if err != nil || saved.Generation != req.Generation {
			problem(w, 500, "无法读取发布记录")
			return
		}
		if err := c.client.Reconcile(r.Context(), &saved, state); err != nil {
			problem(w, 502, "无法核对网关发布记录，请稍后重试")
			return
		}
		if !saved.Current(state) {
			problem(w, 409, "仅可回滚网关当前修订对应的发布")
			return
		}
		desired = saved.Previous
	} else {
		if req.Generation != "" {
			problem(w, 400, "发布请求包含无效的发布记录")
			return
		}
		desired, err = manifest.Decode(req.Manifest)
		if err != nil {
			problem(w, 400, err.Error())
			return
		}
	}
	if err := manifest.Validate(desired, c.plugins); err != nil {
		problem(w, 400, err.Error())
		return
	}
	changes := publisher.Diff(desired, state.Manifest)
	if preview {
		reply(w, 200, map[string]any{"changes": changes, "manifest": desired})
		return
	}
	if len(changes) == 0 {
		problem(w, 400, "没有待发布的变更")
		return
	}
	record, err := publisher.ApplyRelease(r.Context(), c.client, c.dir, actor.Name, req.Generation, state, desired)
	if err != nil {
		code, message := 502, "发布结果待确认，请刷新发布记录核对网关修订"
		var upstream *publisher.ControlError
		switch {
		case record.Generation == "":
			code, message = 500, "无法保存发布记录，请检查 StateDir"
		case record.Status == "applied" || record.Status == "committed":
			code, message = 500, "网关已提交配置，发布记录写入失败，请按发布编号核对"
		case errors.As(err, &upstream) && record.Status == "conflict":
			code, message = 409, "其他发布已生效，请重新载入当前修订"
		case errors.As(err, &upstream) && record.Status == "rejected":
			message = upstream.Message
			if upstream.Status != 401 && upstream.Status != 403 {
				code = upstream.Status
			}
		}
		reply(w, code, map[string]string{"message": message, "generation": record.Generation, "status": record.Status})
		return
	}
	code := http.StatusOK
	if record.Status == "committed" {
		code = http.StatusAccepted
	}
	reply(w, code, map[string]any{"revision": record.AppliedRevision, "generation": record.Generation, "status": record.Status})
}
