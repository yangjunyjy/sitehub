// Package contract describes the file capabilities available to other modules.
package contract

import (
	"context"
	"io"
)

// Settings owns storage configuration and local release validation. Product owns
// publication policy; the file module owns filesystem and object-store access.
type Settings interface {
	EnsureDefaults() error
	GetStorage() (WebsiteStorage, error)
	SaveStorage(WebsiteStorage) error
	ValidateRelease(location string) error
}

type SettingsReader interface {
	GetStorage() (WebsiteStorage, error)
}

// PrivateStorage uses an opaque, encrypted snapshot. A pending reference can be
// committed by blog before external I/O, and retried against the original store.
type PrivateStorage interface {
	PreparePrivate(key string) (snapshot string, err error)
	WritePrivate(context.Context, string, io.Reader, string) error
	ReadPrivate(context.Context, string) (io.ReadCloser, error)
	DeletePrivate(context.Context, string) error
}
