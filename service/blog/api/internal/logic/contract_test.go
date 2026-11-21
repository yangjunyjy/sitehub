package logic

import (
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/types"
	"sitehub/service/blog/domain"
	"strings"
	"testing"
	"time"
)

func TestArticleActionKeepsIDTimeAndOptionalFalse(t *testing.T) {
	at := "2026-11-08T08:00:00.000000123Z"
	flag := false
	input := types.ApplyArticleActionRequest{ID: math.MaxUint64, Version: math.MaxUint64 - 1, RequestID: "operation", Action: "top", IsTop: &flag, ScheduledAt: &at}
	var action domain.ArticleAction
	if err := DecodeRequest(input, &action); err != nil {
		t.Fatal(err)
	}
	if action.ID != input.ID || action.Version != input.Version || action.IsTop == nil || *action.IsTop || action.ScheduledAt == nil || action.ScheduledAt.Format(time.RFC3339Nano) != at {
		t.Fatalf("article action lost precision: %+v", action)
	}
}
func TestArticleBusinessErrorsHaveStableHTTPErrorCodes(t *testing.T) {
	for err, want := range map[error]codes.Code{domain.ErrArticleConflict: codes.Aborted, domain.ErrArticleNotFound: codes.NotFound, domain.ErrArticleForbidden: codes.PermissionDenied, domain.ErrInvalidArticle: codes.InvalidArgument} {
		if got := status.Code(Classify(err)); got != want {
			t.Fatalf("%v => %v, want %v", err, got, want)
		}
	}
}

func TestArticleHTTPResponseKeepsDecimalIDsAndEmptyLists(t *testing.T) {
	reply, err := ReplyAs[payload.ArticlePage](domain.ArticlePage{Items: []domain.ArticleView{{ID: math.MaxUint64, Version: math.MaxUint64 - 1}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(reply)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"id":"18446744073709551615"`, `"version":"18446744073709551614"`, `"tags":[]`, `"tag_ids":[]`, `"total":"0"`, `"category":null`} {
		if !strings.Contains(string(data), field) {
			t.Errorf("missing %s in %s", field, data)
		}
	}
}
