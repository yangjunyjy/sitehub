package console

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"sitehub/gateway/manifest"

	"sitehub/gateway/publisher"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

var generationID = regexp.MustCompile("^[0-9a-f]{24}$")

type historyItem struct {
	Generation       string    `json:"generation"`
	Actor            string    `json:"actor"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	PreviousRevision int64     `json:"previous_revision"`
	AppliedRevision  int64     `json:"applied_revision"`
	RollbackOf       string    `json:"rollback_of,omitempty"`
}

func releaseHistory(dir string) ([]historyItem, error) {
	items := []historyItem{}
	files, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") || !generationID.MatchString(strings.TrimSuffix(file.Name(), ".json")) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		var row historyItem
		if err = json.Unmarshal(data, &row); err != nil {
			return nil, err
		}
		if row.CreatedAt.IsZero() {
			info, err := file.Info()
			if err != nil {
				return nil, err
			}
			row.CreatedAt = info.ModTime()
		}
		items = append(items, row)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

type changeRequest struct {
	Revision   int64           `json:"revision"`
	Manifest   json.RawMessage `json:"manifest"`
	Generation string          `json:"generation"`
}

func (c *console) change(w http.ResponseWriter, r *http.Request, actor *pb.Principal) {
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
	if r.URL.Path == "/api/rollback" {
		if !generationID.MatchString(req.Generation) || len(req.Manifest) > 0 {
			problem(w, 400, "请选择发布记录")
			return
		}
		data, err := os.ReadFile(filepath.Join(c.dir, req.Generation+".json"))
		if errors.Is(err, os.ErrNotExist) {
			problem(w, 404, "找不到这次发布")
			return
		}
		var saved publisher.Release
		if err != nil || json.Unmarshal(data, &saved) != nil {
			problem(w, 500, "无法读取发布记录")
			return
		}
		if saved.Generation != req.Generation || saved.Status != "applied" || saved.AppliedRevision != state.Revision || manifest.Revision(saved.Desired) != manifest.Revision(state.Manifest) {
			problem(w, 409, "仅可回滚当前生效的发布")
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
		reply(w, 200, map[string]any{"changes": changes})
		return
	}
	if len(changes) == 0 {
		problem(w, 400, "没有待发布的变更")
		return
	}
	record, err := publisher.ApplyRelease(r.Context(), c.client, c.dir, actorName(actor), req.Generation, state, desired)
	if err != nil {
		var upstream *publisher.ControlError
		if errors.As(err, &upstream) && upstream.Status == 409 {
			problem(w, 409, "其他发布已生效，请重新载入当前修订")
			return
		}
		problem(w, 502, "发布未确认，请刷新并核对网关修订和发布记录")
		return
	}
	reply(w, 200, map[string]any{"revision": record.AppliedRevision, "generation": record.Generation})
}
