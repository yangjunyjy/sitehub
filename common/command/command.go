// Package command provides the shared Cobra entry point for API, RPC and job services.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sitehub/common/configcenter"
	"sitehub/common/configloader"
	"sitehub/common/controlplane"
	"sitehub/common/governance"
	"sitehub/common/runtimeconfig"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New constructs an independent command. Help and invalid arguments never load
// configuration or initialize database, Redis or RPC connections.
func New[C any](name, defaultConfig string, run func(context.Context, C) error) *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Run the " + name + " service",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			newSettings := func() *viper.Viper {
				settings := viper.New()
				settings.SetConfigFile(configFile)
				// 未改变的空参数不会覆盖go-zero默认值，避免配置文件中缺失的参数被覆盖为零值
				if cmd.Flags().Changed("mode") {
					_ = settings.BindPFlag("Mode", cmd.Flags().Lookup("mode"))
				}
				return settings
			}
			var config C

			prefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
			var bootstrap governance.Config

			// 加载配置
			if err := configloader.Load(newSettings(), prefix, &bootstrap); err != nil {
				return err
			}
			var remote []byte
			var client *configcenter.Proxy
			if bootstrap.ConfigCenter.Enabled {
				if bootstrap.ConfigCenter.Key == "" {
					return fmt.Errorf("ConfigCenter.Key is required")
				}
				var err error
				client, err = controlplane.NewConfigCenter(bootstrap.ControlPlane)
				if err != nil {
					return err
				}
				defer client.Close()
				remote, err = client.Get(cmd.Context(), bootstrap.ConfigCenter.Key)
				if err != nil {
					return err
				}
			}
			if err := configloader.LoadRemote(newSettings(), prefix, &config, remote); err != nil {
				return err
			}
			managed, ok := any(config).(interface{ GovernanceConfig() governance.Config })
			if !ok {
				return run(cmd.Context(), config)
			}
			settings := managed.GovernanceConfig()
			if err := governance.Validate(config, settings); err != nil {
				return err
			}
			source := "local"
			if client != nil {
				source = settings.ControlPlane.Provider + ":" + settings.ConfigCenter.Key
			}
			encoded, _ := json.Marshal(config)
			traffic, err := runtimeconfig.New(settings.Runtime, runtimeconfig.Version(encoded), source)
			if err != nil {
				return err
			}
			runtime, err := governance.New(name, settings, traffic)
			if err != nil {
				return err
			}
			defer runtime.Close()
			if client != nil {
				base := staticVersion(config)
				var reload sync.Mutex
				apply := func(data []byte, watchErr error) {
					reload.Lock()
					defer reload.Unlock()
					version := runtimeconfig.Version(data)
					if watchErr != nil {
						traffic.Reject(version, "configuration center is unavailable")
						return
					}
					if len(data) == 0 {
						traffic.Reject(version, "remote configuration was deleted or empty")
						return
					}
					var candidate C
					if err := configloader.LoadRemote(newSettings(), prefix, &candidate, data); err != nil {
						traffic.Reject(version, "configuration validation failed")
						return
					}
					if staticVersion(candidate) != base {
						traffic.Reject(version, "static configuration changed; restart required")
						return
					}
					value := any(candidate).(interface{ GovernanceConfig() governance.Config }).GovernanceConfig()
					encoded, _ := json.Marshal(candidate)
					if err := traffic.Commit(value.Runtime, runtimeconfig.Version(encoded)); err != nil {
						traffic.Reject(version, err.Error())
					}
				}
				cancel, err := client.Watch(cmd.Context(), settings.ConfigCenter.Key, apply)
				if err != nil {
					return fmt.Errorf("watch remote configuration failed")
				}
				// The provider reads a snapshot and watches from its next revision,
				// so changes between the initial Get and Watch are never lost.
				defer cancel()
			}
			return run(governance.WithContext(cmd.Context(), runtime), config)
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "f", defaultConfig, "configuration file (YAML, JSON or TOML)")
	cmd.Flags().String("mode", "", "override service mode: dev, test, rt, pre or pro")
	return cmd
}

func staticVersion(value any) string {
	encoded, _ := json.Marshal(value)
	var fields map[string]any
	_ = json.Unmarshal(encoded, &fields)
	delete(fields, "Runtime")
	encoded, _ = json.Marshal(fields)
	return runtimeconfig.Version(encoded)
}

// Execute reports a startup error once and returns a nonzero process status.
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
