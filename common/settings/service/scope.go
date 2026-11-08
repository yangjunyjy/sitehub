package services

import (
	"fmt"

	"sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
)

type scopedStore struct {
	base contract.Store
	keys map[string]bool
}

func Scope(base contract.Store, keys ...string) contract.Store {
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	return &scopedStore{base: base, keys: allowed}
}

func (s *scopedStore) check(key string) error {
	if !s.keys[key] {
		return fmt.Errorf("configuration key %q belongs to another module", key)
	}
	return nil
}
func (s *scopedStore) GetConfig(key string) (*settingsmodel.SystemConfig, error) {
	if err := s.check(key); err != nil {
		return nil, err
	}
	return s.base.GetConfig(key)
}
func (s *scopedStore) CreateConfig(row *settingsmodel.SystemConfig) error {
	if err := s.check(row.Key); err != nil {
		return err
	}
	return s.base.CreateConfig(row)
}
func (s *scopedStore) UpdateConfig(row *settingsmodel.SystemConfig) error {
	if err := s.check(row.Key); err != nil {
		return err
	}
	// Check that key and numeric ID identify the same owned row before updating.
	current, err := s.base.GetConfig(row.Key)
	if err != nil {
		return err
	}
	if current == nil || current.ID != row.ID {
		return fmt.Errorf("configuration identity mismatch")
	}
	return s.base.UpdateConfig(row)
}
func (s *scopedStore) BatchUpdateConfigs(rows []contract.BatchUpdateConfig) error {
	// Validate the whole batch before delegating, preserving atomic rejection.
	for _, row := range rows {
		if err := s.check(row.Key); err != nil {
			return err
		}
	}
	return s.base.BatchUpdateConfigs(rows)
}
func (s *scopedStore) UpdateJSONConfigIfUnchanged(key, expected, value string) error {
	if err := s.check(key); err != nil {
		return err
	}
	return s.base.UpdateJSONConfigIfUnchanged(key, expected, value)
}

type scopedAtomicStore struct {
	base contract.AtomicStore
	key  string
}

func ScopeAtomic(base contract.AtomicStore, key string) contract.AtomicStore {
	return &scopedAtomicStore{base: base, key: key}
}
func (s *scopedAtomicStore) GetByKey(key string) (*settingsmodel.SystemConfig, error) {
	if key != s.key {
		return nil, fmt.Errorf("configuration key %q belongs to another module", key)
	}
	return s.base.GetByKey(key)
}
func (s *scopedAtomicStore) MutateValue(row settingsmodel.SystemConfig, change func(string) (string, error)) error {
	if row.Key != s.key {
		return fmt.Errorf("configuration key %q belongs to another module", row.Key)
	}
	return s.base.MutateValue(row, change)
}
