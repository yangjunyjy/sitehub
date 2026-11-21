package main

import (
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
	"sitehub/gateway/publisher"
)

type PublisherConfig struct {
	AdminURL string
	AdminKey string
	StateDir string `json:",default=.tmp/gateway-releases"`
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
			var rollbackOf string
			var expected int64
			if action == "rollback" {
				if from == "" {
					return fmt.Errorf("--release is required")
				}
				data, err := os.ReadFile(from)
				if err != nil {
					return err
				}
				var saved publisher.Release
				if err = json.Unmarshal(data, &saved); err != nil {
					return err
				}
				if saved.Generation == "" || saved.Status != "applied" || saved.AppliedRevision < 1 {
					return fmt.Errorf("rollback requires a confirmed applied native release")
				}
				rollbackOf = saved.Generation
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
			var c PublisherConfig
			if err := configloader.Load(settings, "GATEWAYCTL", &c); err != nil {
				return err
			}
			client, err := c.client()
			if err != nil {
				return err
			}
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
			record, err := publisher.ApplyRelease(cmd.Context(), client, c.StateDir, "gatewayctl", rollbackOf, previous, desired)
			if record.Generation != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "release:", filepath.Join(c.StateDir, record.Generation+".json"), "status:", record.Status)
			}
			return err
		}})
	}
	root.AddCommand(consoleCommand(&file))
	return root
}

func (c PublisherConfig) client() (*publisher.Client, error) {
	u, err := url.Parse(c.AdminURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || len(c.AdminKey) < 32 {
		return nil, fmt.Errorf("valid AdminURL and AdminKey of at least 32 characters are required")
	}
	return &publisher.Client{URL: c.AdminURL, Key: c.AdminKey, HTTP: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
