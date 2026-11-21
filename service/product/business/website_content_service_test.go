package services

import (
	"fmt"
	settings "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	"slices"
	"testing"
)

func TestExistingParticleSelectionsKeepMotionPreferences(t *testing.T) {
	for old, current := range map[string]string{"ring": "artifact", "wave": "cluster", "helix": "terminal"} {
		stored := fmt.Sprintf(`{"home":{"particleShape":%q,"particleAutoplay":false,"particleInteractive":false}}`, old)
		store := &contentReadStore{value: stored}
		content, err := NewWebsiteContentService(store).GetContent()
		if err != nil || content.Home.ParticleShape != current || content.Home.ParticleAutoplay || content.Home.ParticleInteractive || store.value != stored {
			t.Fatalf("read preset %s: %+v, %v", old, content.Home, err)
		}
	}
}

func TestReadMigratesOnlyMatchingWebsiteSeedFields(t *testing.T) {
	defaults := DefaultWebsiteContent()
	if defaults.Home.Title != "部署与运维" || defaults.Home.Description != "" || defaults.Home.Highlight != "" || defaults.Home.PrimaryAction != "下载" || defaults.Home.SecondaryAction != "文档" {
		t.Fatal(defaults.Home)
	}

	store := &contentReadStore{value: `{
        "home":{
            "title":"让部署、运维与", "highlight":"安全管理",
            "description":"由运维团队维护的发行站",
            "primaryAction":"获取 Deploy", "secondaryAction":"阅读手册",
            "capabilityIntro":"项目内设施", "meta":["团队标签"], "cta":"联系项目维护者"
        },
        "footer":{"slogan":"让交付更快、更稳、更安全。","statement":"自定义页脚"}
    }`}
	original := store.value
	service := NewWebsiteContentService(store)
	content, err := service.GetContent()
	if err != nil {
		t.Fatal(err)
	}
	if content.Home.Title != defaults.Home.Title || content.Home.Highlight != "" || content.Home.PrimaryAction != "下载" || content.Footer.Slogan != "" {
		t.Fatalf("old seed still present: %+v", content.Home)
	}
	if content.Home.Description != "由运维团队维护的发行站" || content.Home.SecondaryAction != "阅读手册" || content.Home.CapabilityIntro != "项目内设施" || !slices.Equal(content.Home.Meta, []string{"团队标签"}) || content.Home.CTA != "联系项目维护者" || content.Footer.Statement != "自定义页脚" {
		t.Fatalf("custom copy changed: %+v", content)
	}
	if store.value != original {
		t.Fatal("reading content modified stored settings")
	}
}

// Embedding Store makes any unexpected mutation during a read fail the test.
type contentReadStore struct {
	settings.Store
	value string
}

func (s *contentReadStore) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	return &settingsmodel.SystemConfig{Key: key, Value: s.value}, nil
}
