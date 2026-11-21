package publisher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sitehub/gateway/manifest"

	"time"
)

type Release struct {
	Actor            string            `json:"actor,omitempty"`
	CreatedAt        time.Time         `json:"created_at,omitempty"`
	RollbackOf       string            `json:"rollback_of,omitempty"`
	Generation       string            `json:"generation"`
	Status           string            `json:"status"`
	PreviousRevision int64             `json:"previous_revision"`
	AppliedRevision  int64             `json:"applied_revision"`
	Previous         manifest.Document `json:"previous"`
	Desired          manifest.Document `json:"desired"`
}

func ApplyRelease(ctx context.Context, client *Client, dir, actor, rollbackOf string, previous State, desired manifest.Document) (Release, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Release{}, err
	}
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Release{}, err
	}
	generation := hex.EncodeToString(random[:])
	record := Release{Actor: actor, CreatedAt: time.Now().UTC(), RollbackOf: rollbackOf, Generation: generation, Status: "pending", PreviousRevision: previous.Revision, Previous: previous.Manifest, Desired: desired}
	file := filepath.Join(dir, generation+".json")
	persist := func() error {
		data, _ := json.MarshalIndent(record, "", "  ")
		temporary := file + ".tmp"
		f, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return os.Rename(temporary, file)
	}
	if err := persist(); err != nil {
		return Release{}, err
	}
	applied, err := client.Publish(ctx, desired, previous.Revision)
	if err != nil {
		record.Status = "unconfirmed"
	} else {
		record.Status = "applied"
		record.AppliedRevision = applied.Revision
	}
	if saveErr := persist(); saveErr != nil {
		return record, fmt.Errorf("release result persistence failed; query active revision: %w", saveErr)
	}
	return record, err
}
