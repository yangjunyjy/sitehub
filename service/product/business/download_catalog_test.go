package services

import (
	"errors"
	"testing"

	fileservice "sitehub/service/file/business"
	productcontract "sitehub/service/product/contract"
)

func TestDownloadCatalogRoundTripAndReferenceIntegrity(t *testing.T) {
	store := &downloadConfigStore{value: `{"packages":[],"groups":[]}`}
	service := NewWebsiteContentService(store, nil, fileservice.NewStorageSettings(store))
	catalog := productcontract.WebsiteDownloads{
		Title:  "产品与工具",
		Groups: []productcontract.WebsiteDownloadGroup{{ID: "tools", Name: "运维工具", Description: "独立工具"}, {ID: "deploy", Name: "Deploy"}},
		Packages: []productcontract.WebsiteDownloadPackage{
			{ID: "single", GroupID: "deploy", Product: "Deploy", Edition: "单节点 API 版", Version: "2.0", Platform: "Linux", Status: "published", URL: "https://example.test/single.zip", SortOrder: 20},
			{ID: "distributed", GroupID: "deploy", Product: "Deploy", Edition: "分布式版", Version: "2.1", Platform: "Linux", Status: "draft", SortOrder: 10},
			{ID: "tool", GroupID: "tools", Product: "诊断工具", Version: "1.2", Platform: "跨平台", Status: "published", URL: "https://example.test/tool.zip"},
		},
	}
	if err := service.SaveDownloads(catalog); err != nil {
		t.Fatal(err)
	}
	public, err := service.GetPublicDownloads()
	if err != nil {
		t.Fatal(err)
	}
	if public.Title != catalog.Title || public.Groups[0].ID != "tools" || len(public.Packages) != 2 || public.Packages[0].Edition != "单节点 API 版" || public.Packages[0].SortOrder != 20 {
		t.Fatalf("catalog metadata not preserved: %+v", public)
	}
	if public.Packages[0].URL != "/v1/file/download?id=single" {
		t.Fatal("catalog must retain the protected download entry")
	}
	before := store.value
	catalog.Groups = catalog.Groups[:1] // Still referenced by both published and draft files.
	if err := service.SaveDownloads(catalog); !errors.Is(err, ErrInvalidWebsiteDownloads) {
		t.Fatalf("expected invalid reference error, got %v", err)
	}
	if store.value != before || store.writes != 1 {
		t.Fatal("invalid category deletion changed persisted data")
	}
	catalog.Groups = append(catalog.Groups, catalog.Groups[0])
	if err := service.SaveDownloads(catalog); !errors.Is(err, ErrInvalidWebsiteDownloads) {
		t.Fatalf("duplicate group accepted: %v", err)
	}
}
