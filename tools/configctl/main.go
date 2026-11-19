// Configctl validates and publishes YAML configuration documents to the control plane.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"sitehub/common/controlplane"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type document struct {
	name, key, content string
}

type configStore interface {
	Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Txn(context.Context) clientv3.Txn
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func run(ctx context.Context, args []string, output, errors io.Writer) error {
	flags := flag.NewFlagSet("configctl", flag.ContinueOnError)
	flags.SetOutput(errors)
	dir := flags.String("dir", "deploy/etcd/config", "directory containing YAML configuration documents")
	endpoint := flags.String("endpoint", envDefault("ETCD_ENDPOINT", "127.0.0.1:2379"), "etcd endpoint; defaults to ETCD_ENDPOINT")
	namespace := flags.String("namespace", envDefault("CONTROL_NAMESPACE", "sitehub-dev"), "control-plane namespace")
	group := flags.String("group", envDefault("CONTROL_GROUP", "SITEHUB"), "configuration group")
	prefix := flags.String("prefix", envDefault("CONTROL_PREFIX", "/sitehub"), "control-plane key prefix")
	dialTimeout := flags.Duration("dial-timeout", 5*time.Second, "etcd connection timeout")
	requestTimeout := flags.Duration("request-timeout", 5*time.Second, "timeout per configuration publication")
	overwrite := flags.Bool("overwrite", false, "explicitly replace existing configurations using revision comparisons")
	check := flags.Bool("check", false, "validate files and print destination keys without connecting")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	c := controlplane.Config{
		Provider: "etcd", Endpoints: []string{*endpoint}, Prefix: *prefix,
		Namespace: *namespace, Group: *group,
		Username: os.Getenv("ETCD_USERNAME"), Password: os.Getenv("ETCD_PASSWORD"),
		DialTimeout: *dialTimeout, RequestTimeout: *requestTimeout, LeaseTTL: 15,
	}
	if err := c.Validate(); err != nil {
		return err
	}
	documents, err := loadDocuments(*dir, c.ConfigPrefix())
	if err != nil {
		return err
	}
	if *check {
		for _, d := range documents {
			fmt.Fprintln(output, "valid", d.key)
		}
		return nil
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints: c.Endpoints, Username: c.Username, Password: c.Password, DialTimeout: c.DialTimeout,
		Logger: zap.NewNop(),
	})
	if err != nil {
		return fmt.Errorf("create configuration-store client failed")
	}
	defer client.Close()
	for _, d := range documents {
		request, cancel := context.WithTimeout(ctx, c.RequestTimeout)
		status, err := publish(request, client, d, *overwrite)
		cancel()
		if err != nil {
			return fmt.Errorf("publish %s: %w", d.name, err)
		}
		fmt.Fprintln(output, status, d.key)
	}
	return nil
}

func loadDocuments(dir, prefix string) ([]document, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("invalid seed directory")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("seed directory contains no YAML configuration files")
	}
	sort.Strings(paths)
	documents := make([]document, 0, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read configuration %s failed", filepath.Base(path))
		}
		var value map[string]any
		if err := yaml.Unmarshal(content, &value); err != nil || len(value) == 0 {
			return nil, fmt.Errorf("configuration %s must be a non-empty YAML mapping", filepath.Base(path))
		}
		if strings.Contains(string(content), "${") {
			return nil, fmt.Errorf("configuration %s contains an unresolved environment placeholder", filepath.Base(path))
		}
		name := filepath.Base(path)
		documents = append(documents, document{name: name, key: prefix + name, content: string(content)})
	}
	return documents, nil
}

func publish(ctx context.Context, store configStore, d document, overwrite bool) (string, error) {
	comparison := clientv3.Compare(clientv3.Version(d.key), "=", 0)
	status := "created"
	if overwrite {
		previous, err := store.Get(ctx, d.key)
		if err != nil {
			return "", fmt.Errorf("read previous configuration failed")
		}
		var revision int64
		if len(previous.Kvs) > 0 {
			revision = previous.Kvs[0].ModRevision
			status = "updated"
		}
		comparison = clientv3.Compare(clientv3.ModRevision(d.key), "=", revision)
	}
	response, err := store.Txn(ctx).If(comparison).Then(clientv3.OpPut(d.key, d.content)).Commit()
	if err != nil {
		return "", fmt.Errorf("configuration transaction could not be confirmed; inspect the key before retrying")
	}
	if !response.Succeeded {
		if overwrite {
			return "", fmt.Errorf("configuration changed concurrently; inspect it before publishing again")
		}
		return "kept", nil
	}
	return status, nil
}
