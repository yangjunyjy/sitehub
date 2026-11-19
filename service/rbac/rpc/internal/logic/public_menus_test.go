package logic

import (
	"encoding/json"
	"sitehub/common/protoconv"
	"sitehub/service/rbac/model/entity"
	"strings"
	"testing"
)

func TestPublicMenuProjectionDoesNotExposeAuthorizationFields(t *testing.T) {
	items := publicMenuItems([]*entity.SysMenu{{BaseModel: entity.BaseModel{ID: 9, CreateBy: 77}, SystemCode: "product", MenuName: "Downloads", MenuCode: "private-menu-code", Permission: "secret:permission", Path: "/download", IsVisible: true}})
	value, err := protoconv.Value(items[0])
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(value)
	for _, forbidden := range []string{"permission", "private-menu-code", "createBy", "secret:"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("public response leaked %s", forbidden)
		}
	}
	if !strings.Contains(string(data), `"name":"Downloads"`) || !strings.Contains(string(data), `"systemCode":"product"`) {
		t.Fatalf("public contract missing fields: %s", data)
	}
}
