package services

import (
	"errors"
	"testing"

	"sitehub/service/product/domain"
	"sitehub/service/product/model"
)

type catalogRepository struct {
	repository.ProductDocsRepository
	t         *testing.T
	treeReads int
}

func (r *catalogRepository) Transaction(fn func(repository.ProductDocsRepository) error) error {
	return fn(r)
}
func (*catalogRepository) Site(bool) (domain.DocSite, error) {
	return domain.DocSite{DefaultVersion: "stable"}, nil
}
func (*catalogRepository) Versions() ([]domain.DocVersion, error) {
	return []domain.DocVersion{
		{ID: "stable", ModeID: "server", Status: "published"}, {ID: "next", ModeID: "server", Status: "draft"},
		{ID: "private", ModeID: "internal", Status: "published"},
	}, nil
}
func (*catalogRepository) Modes() ([]domain.DocMode, error) {
	return []domain.DocMode{{ID: "server", Status: "published", DefaultVersion: "stable"}, {ID: "later", Status: "upcoming"}, {ID: "internal", Status: "hidden", DefaultVersion: "private"}}, nil
}
func (r *catalogRepository) Groups(version string) ([]domain.DocGroup, error) {
	if version != "stable" {
		r.t.Fatalf("groups queried for %q", version)
	}
	r.treeReads++
	return []domain.DocGroup{{ID: 7, Title: "入门"}}, nil
}
func (r *catalogRepository) Documents(version string, body bool) ([]domain.ProductDocument, error) {
	if version != "stable" || body {
		r.t.Fatalf("unexpected document query: version=%q body=%t", version, body)
	}
	r.treeReads++
	return []domain.ProductDocument{{GroupID: 7, DocKey: "install"}}, nil
}
func TestCatalogReadsOnlyPublishedMetadataForSelectedVersion(t *testing.T) {
	repo := &catalogRepository{t: t}
	service := websiteContentService{docs: repo}
	catalog, err := service.GetDocCatalog("", "")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ModeID != "server" || len(catalog.Modes) != 2 || len(catalog.Modes[0].VersionIDs) != 1 || catalog.Navigation == nil || len(catalog.Navigation.Groups) != 1 || catalog.Navigation.Groups[0].Items[0].ID != "install" {
		t.Fatalf("unexpected default catalog: %+v", catalog)
	}
	upcoming, err := service.GetDocCatalog("", "later")
	if err != nil || upcoming.Navigation != nil || upcoming.ModeID != "later" {
		t.Fatalf("upcoming mode: %+v, %v", upcoming, err)
	}
	for _, params := range [][2]string{{"next", "server"}, {"private", ""}, {"stable", "later"}, {"", "internal"}, {"invalid", ""}} {
		if _, err := service.GetDocCatalog(params[0], params[1]); !errors.Is(err, ErrWebsiteDocumentNotFound) {
			t.Fatalf("unknown or unpublished mode/version %v: %v", params, err)
		}
	}
	if repo.treeReads != 2 {
		t.Fatalf("expected two scoped metadata queries, got %d", repo.treeReads)
	}
}
