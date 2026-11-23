package control

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sitehub/gateway/manifest"
)

func TestSeedOnlyInitializesAnEmptyStore(t *testing.T) {
	store := &memoryStore{}
	first := staticDocument("first")
	record, err := Seed(context.Background(), store, first)
	if err != nil || record.Revision != 1 || !bytes.Equal(record.Data, manifest.Bytes(first)) {
		t.Fatalf("seed = %+v, %v", record, err)
	}
	record, err = Seed(context.Background(), store, staticDocument("image-upgrade"))
	if err != nil || record.Revision != 1 || !bytes.Equal(record.Data, manifest.Bytes(first)) || store.writes != 1 {
		t.Fatal("startup seed overwrote existing published configuration")
	}
}

func TestConcurrentSeedReturnsTheWinningStoredDocument(t *testing.T) {
	entered, release := make(chan struct{}, 2), make(chan struct{})
	store := &memoryStore{beforeCAS: func() { entered <- struct{}{}; <-release }}
	type outcome struct {
		record Record
		err    error
	}
	results := make(chan outcome, 2)
	for _, value := range []string{"first", "second"} {
		go func(value string) {
			r, err := Seed(context.Background(), store, staticDocument(value))
			results <- outcome{r, err}
		}(value)
	}
	<-entered
	<-entered
	close(release)
	a, b := <-results, <-results
	if a.err != nil || b.err != nil || a.record.Revision != 1 || b.record.Revision != 1 || !bytes.Equal(a.record.Data, b.record.Data) || store.writes != 1 {
		t.Fatalf("concurrent startup did not converge: %+v %+v writes=%d", a, b, store.writes)
	}
}

// A small gRPC protocol fixture exercises the real etcd client and Store
// transaction/watch requests without starting an embedded etcd dependency.
type etcdFixture struct {
	pb.UnimplementedKVServer
	pb.UnimplementedWatchServer
	mu       sync.Mutex
	record   Record
	receipts map[string]Record
	global   int64
	key      string
	starts   chan int64
	events   chan *pb.WatchResponse
	readFail bool
}

func (f *etcdFixture) Range(_ context.Context, req *pb.RangeRequest) (*pb.RangeResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readFail {
		return nil, status.Error(codes.PermissionDenied, "secret backend diagnostic")
	}
	key := string(req.Key)
	if (key != f.key && !strings.HasPrefix(key, f.key+"/releases/")) || req.Serializable || len(req.RangeEnd) != 0 {
		return nil, status.Error(codes.InvalidArgument, "expected linearizable read of the exact configuration key")
	}
	response := &pb.RangeResponse{Header: &pb.ResponseHeader{Revision: f.global}}
	record := f.record
	if key != f.key {
		record = f.receipts[key]
	}
	if record.Revision > 0 {
		response.Kvs = []*mvccpb.KeyValue{{Key: req.Key, Value: bytes.Clone(record.Data), ModRevision: record.Revision}}
	}
	return response, nil
}

func (f *etcdFixture) Txn(_ context.Context, req *pb.TxnRequest) (*pb.TxnResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if (len(req.Compare) != 1 && len(req.Compare) != 2) || len(req.Success) != len(req.Compare) || req.Compare[0].Target != pb.Compare_MOD || req.Compare[0].Result != pb.Compare_EQUAL || string(req.Compare[0].Key) != f.key || req.Success[0].GetRequestPut() == nil || len(req.Failure) != 0 {
		return nil, status.Error(codes.InvalidArgument, "expected atomic mod-revision compare and single whole-document put")
	}
	response := &pb.TxnResponse{Header: &pb.ResponseHeader{Revision: f.global}}
	if req.Compare[0].GetModRevision() != f.record.Revision {
		return response, nil
	}
	put := req.Success[0].GetRequestPut()
	if string(put.Key) != f.key {
		return nil, status.Error(codes.InvalidArgument, "transaction changed a different key")
	}
	if len(req.Compare) == 2 {
		cmp, receipt := req.Compare[1], req.Success[1].GetRequestPut()
		if receipt == nil || cmp.Target != pb.Compare_VERSION || cmp.Result != pb.Compare_EQUAL || cmp.GetVersion() != 0 || !bytes.Equal(cmp.Key, receipt.Key) || !strings.HasPrefix(string(cmp.Key), f.key+"/releases/") {
			return nil, status.Error(codes.InvalidArgument, "invalid create-only release receipt")
		}
		if _, exists := f.receipts[string(cmp.Key)]; exists {
			return response, nil
		}
		f.receipts[string(cmp.Key)] = Record{Revision: f.global + 1, Data: bytes.Clone(receipt.Value)}
	}
	f.global++
	f.record = Record{Revision: f.global, Data: bytes.Clone(put.Value)}
	response.Header.Revision, response.Succeeded = f.global, true
	return response, nil
}

func (f *etcdFixture) Watch(stream pb.Watch_WatchServer) error {
	request, err := stream.Recv()
	if err != nil {
		return err
	}
	create := request.GetCreateRequest()
	if create == nil || string(create.Key) != f.key || len(create.RangeEnd) != 0 {
		return status.Error(codes.InvalidArgument, "expected watch on the exact configuration key")
	}
	f.starts <- create.StartRevision
	f.mu.Lock()
	revision := f.global
	f.mu.Unlock()
	if err := stream.Send(&pb.WatchResponse{Header: &pb.ResponseHeader{Revision: revision}, Created: true, WatchId: 1}); err != nil {
		return err
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case response := <-f.events:
			response.WatchId = 1
			if err := stream.Send(response); err != nil {
				return err
			}
			if response.Canceled {
				return nil
			}
		}
	}
}

func newEtcdFixture(t *testing.T) (*EtcdStore, *etcdFixture) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &etcdFixture{key: "/sitehub/gateway/active", receipts: map[string]Record{}, global: 100, starts: make(chan int64, 8), events: make(chan *pb.WatchResponse, 8)}
	server := grpc.NewServer()
	pb.RegisterKVServer(server, f)
	pb.RegisterWatchServer(server, f)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"http://" + listener.Addr().String()}, DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return &EtcdStore{Client: client, Key: f.key}, f
}

func TestEtcdReceiptCommitsWithConfigAndSurvivesLaterRevisions(t *testing.T) {
	store, _ := newEtcdFixture(t)
	ctx := t.Context()
	id := "0123456789abcdef01234567"
	d := staticDocument("recorded")
	first, err := store.PublishRelease(ctx, 0, manifest.Bytes(d), id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PublishRelease(ctx, first.Revision, manifest.Bytes(staticDocument("duplicate")), id); !errors.Is(err, ErrConflict) {
		t.Fatalf("release ID reused: %v", err)
	}
	if _, err := store.CompareAndSwap(ctx, first.Revision, manifest.Bytes(staticDocument("later"))); err != nil {
		t.Fatal(err)
	}
	got, err := store.Receipt(ctx, id)
	if err != nil || got.Revision != first.Revision || string(got.Data) != fmt.Sprintf("0:%s", manifest.Revision(d)) {
		t.Fatalf("receipt = %v %v", got, err)
	}
	otherID := "000000000000000000000001"
	if _, err := store.PublishRelease(ctx, first.Revision, manifest.Bytes(d), otherID); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale revision accepted: %v", err)
	}
	missing, err := store.Receipt(ctx, otherID)
	if err != nil || missing.Revision != 0 {
		t.Fatalf("stale transaction left a receipt: %+v %v", missing, err)
	}
}

func TestEtcdStoreUsesKeyRevisionAndAtomicCompareAndSwap(t *testing.T) {
	store, fixture := newEtcdFixture(t)
	ctx := context.Background()
	record, err := store.Read(ctx)
	if err != nil || record.Revision != 0 || len(record.Data) != 0 {
		t.Fatalf("missing configuration = %+v, %v", record, err)
	}
	data := manifest.Bytes(staticDocument("first"))
	record, err = store.CompareAndSwap(ctx, 0, data)
	if err != nil || record.Revision != 101 || !bytes.Equal(record.Data, data) {
		t.Fatalf("initial transaction = %+v, %v", record, err)
	}
	fixture.mu.Lock()
	fixture.global = 150 // Writes to unrelated keys do not change our key revision.
	fixture.mu.Unlock()
	record, err = store.Read(ctx)
	if err != nil || record.Revision != 101 {
		t.Fatalf("read used cluster revision rather than key revision: %+v %v", record, err)
	}
	if _, err = store.CompareAndSwap(ctx, 100, manifest.Bytes(staticDocument("stale"))); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale transaction = %v", err)
	}
	record, err = store.Read(ctx)
	if err != nil || !bytes.Equal(record.Data, data) {
		t.Fatal("failed CAS overwrote the active document")
	}
	record, err = store.CompareAndSwap(ctx, 101, manifest.Bytes(staticDocument("second")))
	if err != nil || record.Revision != 151 {
		t.Fatalf("successful commit did not return actual transaction revision: %+v %v", record, err)
	}
	fixture.mu.Lock()
	fixture.readFail = true
	fixture.mu.Unlock()
	if _, err = store.Read(ctx); err == nil || err.Error() != "gateway configuration store unavailable" {
		t.Fatalf("store error leaked backend details: %v", err)
	}
}

func TestEtcdWatchStartsAfterReadRetainsDeletionAndResyncsAfterCompaction(t *testing.T) {
	store, fixture := newEtcdFixture(t)
	fixture.record = Record{Revision: 7, Data: manifest.Bytes(staticDocument("seven"))}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Record, 8)
	failures := make(chan error, 8)
	done := make(chan struct{})
	go func() {
		defer close(done)
		store.Watch(ctx, func(r Record) { updates <- r }, func(err error) { failures <- err })
	}()
	if record := awaitRecord(t, updates); record.Revision != 7 {
		t.Fatalf("initial watch snapshot = %+v", record)
	}
	select {
	case revision := <-fixture.starts:
		if revision != 101 {
			t.Fatalf("watch starts at %d; expected read revision + 1", revision)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not start")
	}
	fixture.events <- &pb.WatchResponse{Header: &pb.ResponseHeader{Revision: 101}, Events: []*mvccpb.Event{{Type: mvccpb.PUT, Kv: &mvccpb.KeyValue{ModRevision: 101, Value: manifest.Bytes(staticDocument("eight"))}}}}
	if record := awaitRecord(t, updates); record.Revision != 101 {
		t.Fatalf("pushed watch snapshot = %+v", record)
	}
	fixture.events <- &pb.WatchResponse{Header: &pb.ResponseHeader{Revision: 102}, Events: []*mvccpb.Event{{Type: mvccpb.DELETE, Kv: &mvccpb.KeyValue{ModRevision: 102}}}}
	select {
	case <-failures:
	case <-time.After(3 * time.Second):
		t.Fatal("deleted active configuration did not report failure")
	}
	select {
	case record := <-updates:
		t.Fatalf("deletion replaced the running snapshot: %+v", record)
	default:
	}
	fixture.mu.Lock()
	fixture.global = 250
	fixture.record = Record{Revision: 112, Data: manifest.Bytes(staticDocument("twelve"))}
	fixture.mu.Unlock()
	fixture.events <- &pb.WatchResponse{Header: &pb.ResponseHeader{Revision: 250}, Canceled: true, CompactRevision: 200}
	if record := awaitRecord(t, updates); record.Revision != 112 {
		t.Fatalf("compacted watch did not reread latest snapshot: %+v", record)
	}
	select {
	case revision := <-fixture.starts:
		if revision != 251 {
			t.Fatalf("resynchronized watch starts at %d", revision)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not resume after compaction")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Watch did not release its loop after cancellation")
	}
}
