// Package client implements file capability ports using the native File RPC.
package client

import (
	"context"
	"io"
	"sitehub/common/protoconv"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	contract "sitehub/service/file/contract"
	pb "sitehub/service/file/rpc/pb/file"
	"time"
)

type Storage struct {
	rpc pb.FileClient
	key string
}

func New(rpc pb.FileClient, key string) *Storage { return &Storage{rpc, key} }
func (s *Storage) call() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return rpcx.Outgoing(ctx, s.key), cancel
}
func (s *Storage) EnsureDefaults() error {
	ctx, cancel := s.call()
	defer cancel()
	_, err := s.rpc.EnsureStorage(ctx, &pb.Empty{})
	return err
}
func (s *Storage) GetStorage() (contract.WebsiteStorage, error) {
	ctx, cancel := s.call()
	defer cancel()
	out, err := s.rpc.StorageSnapshot(ctx, &pb.Empty{})
	var value contract.WebsiteStorage
	if err == nil {
		err = protoconv.Decode(out, &value)
	}
	return value, err
}
func (s *Storage) SaveStorage(value contract.WebsiteStorage) error {
	ctx, cancel := s.call()
	defer cancel()
	in := &pb.WebsiteStorage{}
	if err := protoconv.Encode(value, in); err != nil {
		return err
	}
	_, err := s.rpc.SaveStorage(ctx, in)
	return err
}
func (s *Storage) ValidateRelease(location string) error {
	ctx, cancel := s.call()
	defer cancel()
	_, err := s.rpc.ValidateRelease(ctx, &pb.LocationRequest{Location: location})
	return err
}
func (s *Storage) PreparePrivate(key string) (string, error) {
	ctx, cancel := s.call()
	defer cancel()
	out, err := s.rpc.PreparePrivate(ctx, &pb.SnapshotRequest{Key: key})
	if err != nil {
		return "", err
	}
	return out.Snapshot, nil
}
func (s *Storage) WritePrivate(ctx context.Context, snapshot string, body io.Reader, mime string) error {
	// The blog asset input is seekable, allowing a verified length without buffering.
	seeker, ok := body.(io.Seeker)
	if !ok {
		return io.ErrUnexpectedEOF
	}
	position, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	end, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err = seeker.Seek(position, io.SeekStart); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := s.rpc.WritePrivate(rpcx.Outgoing(ctx, s.key))
	if err != nil {
		return err
	}
	if err = stream.Send(&pb.StreamChunk{Snapshot: snapshot, Mime: mime, Size: end - position}); err != nil {
		return err
	}
	if err = streamio.Copy(ctx, body, func(data []byte) error { return stream.Send(&pb.StreamChunk{Data: data}) }); err != nil {
		return err
	}
	_, err = stream.CloseAndRecv()
	return err
}

type reader struct {
	recv   func() (*pb.StreamChunk, error)
	cancel context.CancelFunc
	buffer []byte
}

func (r *reader) Read(p []byte) (int, error) {
	for len(r.buffer) == 0 {
		chunk, err := r.recv()
		if err != nil {
			return 0, err
		}
		r.buffer = chunk.Data
	}
	n := copy(p, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}
func (r *reader) Close() error { r.cancel(); return nil }
func (s *Storage) ReadPrivate(ctx context.Context, snapshot string) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)
	stream, err := s.rpc.ReadPrivate(rpcx.Outgoing(ctx, s.key), &pb.SnapshotRequest{Snapshot: snapshot})
	if err != nil {
		cancel()
		return nil, err
	}
	return &reader{recv: stream.Recv, cancel: cancel}, nil
}
func (s *Storage) DeletePrivate(ctx context.Context, snapshot string) error {
	_, err := s.rpc.DeletePrivate(rpcx.Outgoing(ctx, s.key), &pb.SnapshotRequest{Snapshot: snapshot})
	return err
}

var _ contract.PrivateStorage = (*Storage)(nil)
var _ contract.Settings = (*Storage)(nil)
