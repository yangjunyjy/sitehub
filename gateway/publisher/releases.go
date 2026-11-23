package publisher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"sitehub/gateway/manifest"
)

var GenerationID = regexp.MustCompile("^[0-9a-f]{24}$")

type ReleaseSummary struct {
	Actor            string    `json:"actor,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	RollbackOf       string    `json:"rollback_of,omitempty"`
	Generation       string    `json:"generation"`
	Status           string    `json:"status"`
	Message          string    `json:"message,omitempty"`
	PreviousRevision int64     `json:"previous_revision"`
	AppliedRevision  int64     `json:"applied_revision"`
}

type Release struct {
	ReleaseSummary
	Previous manifest.Document `json:"previous"`
	Desired  manifest.Document `json:"desired"`
}

func LoadRelease(file string) (Release, error) {
	var r Release
	data, err := os.ReadFile(file)
	if err != nil {
		return r, err
	}
	if err = json.Unmarshal(data, &r); err != nil {
		return r, err
	}
	if !GenerationID.MatchString(r.Generation) || r.PreviousRevision < 1 || r.Status == "" || r.Previous.Version == 0 || r.Desired.Version == 0 {
		return r, errors.New("invalid release record")
	}
	return r, nil
}

func (r ReleaseSummary) NeedsReceipt() bool {
	return r.Status == "pending" || r.Status == "unconfirmed" || r.Status == "committed"
}

func (r Release) Current(state State) bool {
	return (r.Status == "applied" || r.Status == "committed") && r.AppliedRevision == state.Revision && manifest.Revision(r.Desired) == manifest.Revision(state.Manifest)
}

// Reconcile consults the durable transaction receipt, including after a restart.
// A missing receipt cannot disprove an in-flight commit. Legacy unidentified
// writes remain unknown; an identical manifest alone does not prove authorship.
func (c *Client) Reconcile(ctx context.Context, r *Release, state State) error {
	if !r.NeedsReceipt() {
		return nil
	}
	if !GenerationID.MatchString(r.Generation) {
		return fmt.Errorf("invalid release ID")
	}
	var outcome struct {
		Revision int64
		Receipt  string
	}
	err := c.api(ctx, http.MethodGet, "/v1/releases/"+r.Generation, 0, "", nil, &outcome)
	var control *ControlError
	if errors.As(err, &control) && control.Status == 404 {
		return nil
	}
	if err != nil {
		return err
	}
	if outcome.Revision <= r.PreviousRevision || outcome.Receipt != fmt.Sprintf("%d:%s", r.PreviousRevision, manifest.Revision(r.Desired)) {
		return fmt.Errorf("release receipt does not match recorded configuration")
	}
	r.AppliedRevision = outcome.Revision
	r.Status = "committed"
	r.Message = ""
	if state.AppliedRevision == outcome.Revision {
		r.Status = "applied"
	}
	return nil
}

func persistRelease(file string, record Release) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(file), ".release-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Chtimes(f.Name(), record.CreatedAt, record.CreatedAt); err != nil {
		return err
	}
	return os.Rename(f.Name(), file)
}

func ApplyRelease(ctx context.Context, client *Client, dir, actor, rollbackOf string, previous State, desired manifest.Document) (Release, error) {
	if previous.FileManaged {
		return Release{}, fmt.Errorf("file-managed gateway requires file reload")
	}
	if previous.DesiredUnavailable || previous.Revision < 1 {
		return Release{}, fmt.Errorf("publication requires the current stored revision")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Release{}, err
	}
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Release{}, err
	}
	record := Release{ReleaseSummary: ReleaseSummary{
		Actor: actor, CreatedAt: time.Now().UTC(), RollbackOf: rollbackOf, Generation: hex.EncodeToString(random[:]), Status: "pending", PreviousRevision: previous.Revision,
	}, Previous: previous.Manifest, Desired: desired}
	file := filepath.Join(dir, record.Generation+".json")
	if err := persistRelease(file, record); err != nil {
		return Release{}, err
	}
	applied, err := client.request(ctx, http.MethodPut, previous.Revision, record.Generation, desired)
	if err == nil {
		record.AppliedRevision = applied.Revision
		record.Status = "committed"
		record.Message = applied.LastError
		if applied.AppliedRevision == applied.Revision {
			record.Status = "applied"
		}
	} else {
		record.Message = err.Error()
		var control *ControlError
		switch {
		case errors.As(err, &control) && control.Status == 409:
			record.Status = "conflict"
		case errors.As(err, &control) && (control.Outcome == "rejected" || control.Status >= 400 && control.Status < 500 && control.Status != 408):
			record.Status = "rejected"
		default:
			record.Status = "unconfirmed"
		}
	}
	if saveErr := persistRelease(file, record); saveErr != nil {
		return record, errors.Join(err, fmt.Errorf("release result could not be saved; check release %s: %w", record.Generation, saveErr))
	}
	return record, err
}
