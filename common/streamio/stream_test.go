package streamio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type chunk struct {
	data []byte
	size int64
}

func (c chunk) GetData() []byte { return c.data }
func (c chunk) GetSize() int64  { return c.size }
func TestUploadRejectsIncompleteAndOversizedBodies(t *testing.T) {
	for _, body := range [][]byte{[]byte("short"), []byte("too-long-content")} {
		calls := 0
		file, _, err := Receive(context.Background(), func() (chunk, error) {
			calls++
			if calls == 1 {
				return chunk{body, 10}, nil
			}
			return chunk{}, io.EOF
		}, 20)
		if file != nil || status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid upload accepted: %v", err)
		}
	}
}
func TestRemoteSeekerSupportsRangesAndHead(t *testing.T) {
	data := []byte("0123456789")
	opens := 0
	for _, method := range []string{"GET", "HEAD"} {
		reader := &Seeker{Size: int64(len(data)), Open: func(offset, length int64) (io.ReadCloser, error) {
			opens++
			return io.NopCloser(bytes.NewReader(data[offset : offset+length])), nil
		}}
		request := httptest.NewRequest(method, "/download", nil)
		if method == "GET" {
			request.Header.Set("Range", "bytes=2-5")
		}
		response := httptest.NewRecorder()
		response.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(response, request, "app.tgz", time.Time{}, reader)
		reader.Close()
		if method == "GET" && (response.Code != 206 || response.Body.String() != "2345") {
			t.Fatal(response.Code, response.Body.String())
		}
		if method == "HEAD" && (response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") != "10") {
			t.Fatal("invalid HEAD response")
		}
	}
	if opens != 1 {
		t.Fatalf("HEAD unexpectedly downloaded the body: %d streams", opens)
	}
}
