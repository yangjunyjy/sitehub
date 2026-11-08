package logic

import (
	"math"
	"sitehub/common/protoconv"
	domain "sitehub/service/blog/domain"
	pb "sitehub/service/blog/rpc/pb/blog"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestArticleContractPreservesIDsOptionalFalseAndTime(t *testing.T) {
	flag := false
	at := time.Date(2026, 11, 8, 8, 0, 0, 123, time.UTC)
	input := domain.ArticleAction{ID: math.MaxUint64, Version: math.MaxUint64 - 1, RequestID: "operation", Action: "top", IsTop: &flag, ScheduledAt: &at}
	message := &pb.ArticleAction{}
	if err := protoconv.Encode(input, message); err != nil {
		t.Fatal(err)
	}
	var decoded domain.ArticleAction
	if err := protoconv.Decode(message, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != input.ID || decoded.Version != input.Version || decoded.IsTop == nil || *decoded.IsTop || decoded.ScheduledAt == nil || !decoded.ScheduledAt.Equal(at) {
		t.Fatalf("contract lost precision or presence: %+v", decoded)
	}
}
func TestArticleBusinessErrorsHaveStableRPCStatus(t *testing.T) {
	for err, want := range map[error]codes.Code{domain.ErrArticleConflict: codes.Aborted, domain.ErrArticleNotFound: codes.NotFound, domain.ErrArticleForbidden: codes.PermissionDenied, domain.ErrInvalidArticle: codes.InvalidArgument} {
		if got := status.Code(classify(err)); got != want {
			t.Fatalf("%v => %v, want %v", err, got, want)
		}
	}
}
