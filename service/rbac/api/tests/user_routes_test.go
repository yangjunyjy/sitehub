package api_test

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/router"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"sitehub/common/apiresponse"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/api/internal/config"
	"sitehub/service/rbac/api/internal/handler"
	"sitehub/service/rbac/api/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

type userRPC struct {
	pb.RbacClient
	id       uint64
	metadata metadata.MD
	calls    int
}

func (client *userRPC) GetUser(ctx context.Context, req *pb.IDRequest, _ ...grpc.CallOption) (*pb.EntitySysUser, error) {
	client.calls++
	client.id = req.Id
	client.metadata, _ = metadata.FromOutgoingContext(ctx)
	return &pb.EntitySysUser{Id: req.Id}, nil
}

func TestUserRoutePreservesIdentityAndFullWidthID(t *testing.T) {
	const key = "internal-service-key-for-the-test"
	apiresponse.Install()
	server, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}
	client := &userRPC{}
	handler.RegisterHandlers(server, &svc.ServiceContext{RPC: client, Config: config.Config{ServiceKey: key}})
	mux := router.NewRouter()
	for _, route := range server.Routes() {
		if err := mux.Handle(route.Method, route.Path, rpcx.HTTPContext(route.Handler)); err != nil {
			t.Fatal(err)
		}
	}
	id := strconv.FormatUint(math.MaxUint64, 10)
	req := httptest.NewRequest(http.MethodGet, "/v1/rbac/users/detail?id="+id, nil)
	req.Header.Set("Authorization", "Bearer user-session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	var payload struct{ Data struct{ ID string } }
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || payload.Data.ID != id || client.id != math.MaxUint64 {
		t.Fatalf("status=%d response ID=%q RPC ID=%d", response.Code, payload.Data.ID, client.id)
	}
	for name, want := range map[string]string{"authorization": "user-session", "x-service-key": key} {
		if got := client.metadata.Get(name); len(got) != 1 || got[0] != want {
			t.Errorf("RPC metadata %s = %v", name, got)
		}
	}
	invalid := httptest.NewRecorder()
	mux.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/v1/rbac/users/detail?id=invalid", nil))
	if invalid.Code != http.StatusBadRequest || client.calls != 1 {
		t.Fatalf("invalid id: status=%d RPC calls=%d", invalid.Code, client.calls)
	}
}
