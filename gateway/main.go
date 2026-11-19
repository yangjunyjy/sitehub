package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sitehub/common/command"
	"sitehub/common/configloader"
	"sitehub/gateway/internal/app"
	"sitehub/gateway/internal/config"
	"sitehub/gateway/plugin"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	command.Execute(newCommand())
}
func newCommand() *cobra.Command {
	return newCommandWithRegistry(plugin.DefaultRegistry())
}

func newCommandWithRegistry(plugins *plugin.Registry) *cobra.Command {
	var file, listen string
	load := func(cmd *cobra.Command) (config.Config, error) {
		settings := viper.New()
		settings.SetConfigFile(file)
		if cmd.Flags().Changed("listen") {
			_ = settings.BindPFlag("Listen", cmd.Flags().Lookup("listen"))
		}
		var c config.Config
		err := configloader.Load(settings, "GATEWAY", &c)
		if err == nil {
			if !filepath.IsAbs(c.Manifest) {
				c.Manifest = filepath.Join(filepath.Dir(file), c.Manifest)
			}
			for _, value := range []*string{&c.TLS.Certificate, &c.TLS.Key} {
				if *value != "" && !filepath.IsAbs(*value) {
					*value = filepath.Join(filepath.Dir(file), *value)
				}
			}
			err = c.Validate()
		}
		return c, err
	}
	root := &cobra.Command{Use: "gateway", Short: "Gateway with declarative proxy configuration and configurable plugins", SilenceErrors: true, SilenceUsage: true, Args: cobra.NoArgs}
	root.PersistentFlags().StringVarP(&file, "config", "f", "gateway/etc/gateway.yaml", "gateway process configuration")
	root.PersistentFlags().StringVar(&listen, "listen", "", "override public listener")
	root.RunE = func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return app.RunWithRegistry(ctx, c, plugins)
	}
	root.AddCommand(&cobra.Command{Use: "check", Short: "Validate process settings, included routes and plugin parameters without connecting to services", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		if _, err = app.Check(c, plugins); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "gateway configuration, proxy routes and plugins are valid")
		return nil
	}})
	root.AddCommand(&cobra.Command{Use: "reload", Short: "Atomically reload the running gateway's proxy files", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		if c.SourceType() != "file" {
			return fmt.Errorf("reload requires Source: file; use gatewayctl publish for etcd")
		}
		if _, err = app.ReadRoutes(c, plugins); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+adminAddress(c.Admin.Listen)+"/v1/reload", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Admin.Key)
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("gateway reload request failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gateway rejected reload (HTTP %d); previous configuration remains active; inspect control status", resp.StatusCode)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "gateway proxy configuration reloaded")
		return nil
	}})
	root.AddCommand(&cobra.Command{Use: "health", Short: "Check local gateway readiness", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		address := adminAddress(c.Admin.Listen)
		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+address+"/readyz", nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("gateway is not ready")
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("gateway is not ready")
		}
		return nil
	}})
	return root
}

func adminAddress(address string) string {
	if host, port, err := net.SplitHostPort(address); err == nil {
		if host == "0.0.0.0" || host == "" {
			return net.JoinHostPort("127.0.0.1", port)
		}
		if host == "::" {
			return net.JoinHostPort("::1", port)
		}
	}
	return address
}
