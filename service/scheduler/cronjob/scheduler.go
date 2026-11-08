package main

import (
	"context"
	"fmt"
	"sitehub/common/command"
	"sitehub/common/governance"
	"sitehub/service/scheduler/cronjob/internal/config"
	"sitehub/service/scheduler/cronjob/internal/svc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	command.Execute(command.New("scheduler", "etc/scheduler.yaml", run))
}

func run(parent context.Context, c config.Config) error {
	if len(c.ServiceKey) < 32 {
		return fmt.Errorf("ServiceKey must contain at least 32 characters")
	}
	if err := c.SetUp(); err != nil {
		return err
	}
	ctx, err := svc.NewServiceContext(c, governance.FromContext(parent).Traffic)
	if err != nil {
		return err
	}
	shutdown, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx.Jobs.Start()
	governance.FromContext(parent).Ready()
	<-shutdown.Done()
	governance.FromContext(parent).Withdraw()
	ctx.Jobs.Stop()
	return nil
}
