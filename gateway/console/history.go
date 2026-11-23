package console

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"sitehub/gateway/publisher"
)

func (c *console) history(w http.ResponseWriter, r *http.Request, _ account) {
	page, limit := 1, 20
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
	}
	if err != nil || page < 1 {
		problem(w, 400, "页码必须是正整数")
		return
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
	}
	if err != nil || limit < 1 || limit > 100 {
		problem(w, 400, "每页条数须在 1 至 100 之间")
		return
	}
	items, total, err := c.historyPage(r.Context(), page, limit, nil)
	if err != nil {
		problem(w, 500, "无法读取发布记录，请检查 StateDir")
		return
	}
	reply(w, 200, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (c *console) historyPage(ctx context.Context, page, limit int, state *publisher.State) ([]publisher.ReleaseSummary, int, error) {
	items := []publisher.ReleaseSummary{}
	files, err := os.ReadDir(c.dir)
	if errors.Is(err, os.ErrNotExist) {
		return items, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	type entry struct {
		id      string
		created time.Time
	}
	var records []entry
	for _, file := range files {
		id, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok || !publisher.GenerationID.MatchString(id) || !file.Type().IsRegular() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			return nil, 0, err
		}
		records = append(records, entry{id, info.ModTime()})
	}
	total := len(records)
	if total == 0 || page-1 > (total-1)/limit {
		return items, total, nil
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].created.Equal(records[j].created) {
			return records[i].id > records[j].id
		}
		return records[i].created.After(records[j].created)
	})
	// A slow or unavailable control store must not hide the rest of the journal.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var snapshotErr error
	start := (page - 1) * limit
	for _, file := range records[start:min(total, start+limit)] {
		record, err := publisher.LoadRelease(filepath.Join(c.dir, file.id+".json"))
		if err != nil || record.Generation != file.id {
			items = append(items, publisher.ReleaseSummary{Generation: file.id, CreatedAt: file.created, Status: "corrupt", Message: "发布记录无法解析"})
			continue
		}
		if record.CreatedAt.IsZero() {
			record.CreatedAt = file.created
		}
		if record.NeedsReceipt() {
			if state == nil && snapshotErr == nil {
				var snapshot publisher.State
				snapshot, snapshotErr = c.client.Snapshot(ctx)
				state = &snapshot
			}
			if snapshotErr != nil || c.client.Reconcile(ctx, &record, *state) != nil {
				record.Message = "暂时无法核对发布结果"
			}
		}
		items = append(items, record.ReleaseSummary)
	}
	return items, total, nil
}
