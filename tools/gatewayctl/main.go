package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sitehub/common/command"
	"sitehub/common/configloader"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
	"sitehub/tools/gatewayctl/internal/publisher"
)

type config struct {
	AdminURL string
	AdminKey string
	StateDir string `json:",default=.tmp/gateway-releases"`
}
type release struct {
	Generation       string            `json:"generation"`
	Status           string            `json:"status"`
	PreviousRevision int64             `json:"previous_revision"`
	AppliedRevision  int64             `json:"applied_revision"`
	Previous         manifest.Document `json:"previous"`
	Desired          manifest.Document `json:"desired"`
}

func main() { command.Execute(newCommand()) }
func newCommand() *cobra.Command {
	var file, manifestFile, from string
	root := &cobra.Command{Use: "gatewayctl", Short: "Validate, diff, atomically publish and roll back gateway configurations", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringVarP(&file, "config", "f", "gateway/etc/gatewayctl.yaml", "Viper configuration file")
	root.PersistentFlags().StringVar(&manifestFile, "manifest", "gateway/etc/routes.json", "native declarative manifest")
	root.PersistentFlags().StringVar(&from, "release", "", "applied release whose previous content will be restored")
	for _, action := range []string{"validate", "diff", "publish", "rollback"} {
		action := action
		root.AddCommand(&cobra.Command{Use: action, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			var desired manifest.Document
			var expected int64
			if action == "rollback" {
				if from == "" {
					return fmt.Errorf("--release is required")
				}
				data, err := os.ReadFile(from)
				if err != nil {
					return err
				}
				var saved release
				if err = json.Unmarshal(data, &saved); err != nil {
					return err
				}
				if saved.Generation == "" || saved.Status != "applied" || saved.AppliedRevision < 1 {
					return fmt.Errorf("rollback requires a confirmed applied native release")
				}
				desired = saved.Previous
				expected = saved.AppliedRevision
				if err = manifest.Validate(desired, plugin.DefaultRegistry()); err != nil {
					return err
				}
			} else {
				var err error
				desired, err = manifest.Read(manifestFile)
				if err != nil {
					return err
				}
			}
			if action == "validate" {
				fmt.Fprintln(cmd.OutOrStdout(), "native gateway manifest and plugin schemas valid")
				return nil
			}
			settings := viper.New()
			settings.SetConfigFile(file)
			var c config
			if err := configloader.Load(settings, "GATEWAYCTL", &c); err != nil {
				return err
			}
			u, err := url.Parse(c.AdminURL)
			if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || len(c.AdminKey) < 32 {
				return fmt.Errorf("valid AdminURL and AdminKey of at least 32 characters are required")
			}
			client := &publisher.Client{URL: c.AdminURL, Key: c.AdminKey, HTTP: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
			previous, err := client.Snapshot(cmd.Context())
			if err != nil {
				return err
			}
			if expected != 0 && previous.Revision != expected {
				return fmt.Errorf("rollback refused: another release has replaced this version; review current configuration first")
			}
			for _, change := range publisher.Diff(desired, previous.Manifest) {
				fmt.Fprintln(cmd.OutOrStdout(), change)
			}
			if action == "diff" {
				return nil
			}
			if err = os.MkdirAll(c.StateDir, 0700); err != nil {
				return err
			}
			var random [12]byte
			if _, err = rand.Read(random[:]); err != nil {
				return err
			}
			generation := hex.EncodeToString(random[:])
			record := release{Generation: generation, Status: "pending", PreviousRevision: previous.Revision, Previous: previous.Manifest, Desired: desired}
			file := filepath.Join(c.StateDir, generation+".json")
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
			if err = persist(); err != nil {
				return err
			}
			applied, err := client.Publish(cmd.Context(), desired, previous.Revision)
			if err != nil {
				record.Status = "unconfirmed"
			} else {
				record.Status = "applied"
				record.AppliedRevision = applied.Revision
			}
			if saveErr := persist(); saveErr != nil {
				return fmt.Errorf("release result persistence failed; query active revision: %w", saveErr)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "release:", file, "status:", record.Status)
			return err
		}})
	}
	return root
}
