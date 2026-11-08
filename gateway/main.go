package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sitehub/common/command"
	"sitehub/common/configloader"
	"sitehub/gateway/internal/app"
	"sitehub/gateway/internal/config"
)

func main() { command.Execute(newCommand()) }
func newCommand() *cobra.Command {
	var file, listen string
	load := func(cmd *cobra.Command) (config.Config, error) {
		settings := viper.New()
		settings.SetConfigFile(file)
		_ = settings.BindPFlag("Listen", cmd.Flags().Lookup("listen"))
		var c config.Config
		err := configloader.Load(settings, "GATEWAY", &c)
		if err == nil {
			err = c.Validate()
		}
		return c, err
	}
	root := &cobra.Command{Use: "gateway", Short: "Native Go gateway with atomic configuration and Nacos discovery", SilenceErrors: true, SilenceUsage: true, Args: cobra.NoArgs}
	root.PersistentFlags().StringVarP(&file, "config", "f", "gateway/etc/gateway.yaml", "Viper bootstrap configuration")
	root.PersistentFlags().StringVar(&listen, "listen", "", "override public listener")
	root.RunE = func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return app.Run(ctx, c)
	}
	root.AddCommand(&cobra.Command{Use: "health", Short: "Check local gateway readiness", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := load(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		address := c.Admin.Listen
		if len(address) > 8 && address[:8] == "0.0.0.0:" {
			address = "127.0.0.1:" + address[8:]
		}
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
