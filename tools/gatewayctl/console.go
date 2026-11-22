package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"sitehub/common/configloader"
	"sitehub/gateway/console"
	"sitehub/gateway/plugin"
	"syscall"
	"time"
)

func consoleCommand(file *string) *cobra.Command {
	return &cobra.Command{Use: "console", Short: "Serve the standalone gateway administration console", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		settings := viper.New()
		settings.SetConfigFile(*file)
		var cfg struct {
			PublisherConfig
			console.Config
		}
		if err := configloader.Load(settings, "GATEWAYCTL", &cfg); err != nil {
			return err
		}
		client, err := cfg.client()
		if err != nil {
			return err
		}
		handler, err := console.New(cfg.Config, client, plugin.DefaultRegistry(), cfg.StateDir)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		server := &http.Server{Addr: cfg.Config.Listen, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 40 * time.Second, IdleTimeout: 60 * time.Second}
		stopped := make(chan error, 1)
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			stopped <- server.Shutdown(shutdown)
		}()
		fmt.Fprintln(cmd.OutOrStdout(), "Gateway console listening on", cfg.Config.Listen)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return <-stopped
	}}
}
