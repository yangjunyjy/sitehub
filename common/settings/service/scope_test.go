package services

import (
	"errors"
	"testing"

	"sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
)

type scopeFixture struct {
	contract.Store
	writes int
	cause  error
}

func (s *scopeFixture) BatchUpdateConfigs([]contract.BatchUpdateConfig) error { s.writes++; return nil }
func (s *scopeFixture) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	return &settingsmodel.SystemConfig{Key: key}, s.cause
}
func (s *scopeFixture) UpdateJSONConfigIfUnchanged(key, expected, value string) error { return s.cause }

func TestConfigOwnershipRejectsWholeCrossModuleBatch(t *testing.T) {
	base := &scopeFixture{}
	product := Scope(base, "website.content", "website.downloads")
	if _, err := product.GetConfig("website.storage"); err == nil {
		t.Fatal("product read file secrets")
	}
	err := product.BatchUpdateConfigs([]contract.BatchUpdateConfig{{Key: "website.content"}, {Key: "website.storage"}})
	if err == nil || base.writes != 0 {
		t.Fatal("cross-module batch partially wrote data")
	}
	if err = product.BatchUpdateConfigs([]contract.BatchUpdateConfig{{Key: "website.content"}}); err != nil || base.writes != 1 {
		t.Fatalf("owned write rejected: %v", err)
	}
	base.cause = errors.New("database offline")
	if err = product.UpdateJSONConfigIfUnchanged("website.downloads", "old", "new"); !errors.Is(err, base.cause) {
		t.Fatalf("dependency failure was hidden: %v", err)
	}
}
