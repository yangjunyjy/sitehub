package contract

import (
	settingsmodel "sitehub/common/settings/model"
)

// Store is the transitional configuration port. Each module receives a view
// restricted to its own keys while phase one retains the existing SQL table.
type Store interface {
	GetConfig(string) (*settingsmodel.SystemConfig, error)
	CreateConfig(*settingsmodel.SystemConfig) error
	UpdateConfig(*settingsmodel.SystemConfig) error
	BatchUpdateConfigs([]BatchUpdateConfig) error
	UpdateJSONConfigIfUnchanged(key, expected, value string) error
}

type AtomicStore interface {
	GetByKey(string) (*settingsmodel.SystemConfig, error)
	MutateValue(settingsmodel.SystemConfig, func(string) (string, error)) error
}
