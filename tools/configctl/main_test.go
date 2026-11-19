package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestPublishCreatesOnlyMissingConfigurations(t *testing.T) {
	store := &fakeStore{success: false}
	status, err := publish(context.Background(), store, document{key: "/config/test.yaml", content: "Timeout: 1\n"}, false)
	if err != nil || status != "kept" || store.reads != 0 {
		t.Fatalf("status=%s err=%v reads=%d", status, err, store.reads)
	}
	compare := etcdserverpb.Compare(store.comparison)
	if compare.Target != etcdserverpb.Compare_VERSION || compare.GetVersion() != 0 {
		t.Fatalf("default publication must compare Version(key)=0, got %v", compare)
	}
}

func TestOverwriteComparesTheObservedRevision(t *testing.T) {
	store := &fakeStore{success: true, previous: &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{ModRevision: 42}}}}
	status, err := publish(context.Background(), store, document{key: "/config/test.yaml", content: "Timeout: 1\n"}, true)
	compare := etcdserverpb.Compare(store.comparison)
	if err != nil || status != "updated" || compare.Target != etcdserverpb.Compare_MOD || compare.GetModRevision() != 42 {
		t.Fatalf("status=%s err=%v compare=%v", status, err, compare)
	}
	store.success = false
	if _, err := publish(context.Background(), store, document{key: "/config/test.yaml"}, true); err == nil {
		t.Fatal("concurrent configuration replacement was not rejected")
	}
}

func TestSeedValidationHappensBeforePublication(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rbac-api.yaml")
	if err := os.WriteFile(path, []byte("Timeout: 30000\n"), 0600); err != nil {
		t.Fatal(err)
	}
	documents, err := loadDocuments(dir, "/sitehub/config/test/SITEHUB/")
	if err != nil || len(documents) != 1 || documents[0].key != "/sitehub/config/test/SITEHUB/rbac-api.yaml" {
		t.Fatalf("documents=%v err=%v", documents, err)
	}
	for _, invalid := range []string{"", "[]\n", "Secret: ${ENV_VALUE}\n", "Key: [\n"} {
		if err := os.WriteFile(path, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadDocuments(dir, "/config/"); err == nil {
			t.Fatalf("accepted invalid seed %q", invalid)
		}
	}
}

type fakeStore struct {
	previous   *clientv3.GetResponse
	reads      int
	success    bool
	comparison clientv3.Cmp
}

func (f *fakeStore) Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	f.reads++
	return f.previous, nil
}
func (f *fakeStore) Txn(context.Context) clientv3.Txn { return &fakeTxn{store: f} }

type fakeTxn struct{ store *fakeStore }

func (f *fakeTxn) If(comparisons ...clientv3.Cmp) clientv3.Txn {
	f.store.comparison = comparisons[0]
	return f
}
func (f *fakeTxn) Then(...clientv3.Op) clientv3.Txn { return f }
func (f *fakeTxn) Else(...clientv3.Op) clientv3.Txn { return f }
func (f *fakeTxn) Commit() (*clientv3.TxnResponse, error) {
	return &clientv3.TxnResponse{Succeeded: f.store.success}, nil
}
