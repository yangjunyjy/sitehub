// Package streamio provides bounded-memory transport for RPC file streams.
package streamio

import (
	"context"
	"errors"
	"io"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Chunk interface {
	GetData() []byte
	GetSize() int64
}

// Receive spools one upload with a verified declared size. The caller owns the
// successful temporary file; every failed or canceled upload is removed here.
func Receive[T Chunk](ctx context.Context, recv func() (T, error), limit int64) (file *os.File, metadata T, err error) {
	metadata, err = recv()
	if err != nil {
		return nil, metadata, err
	}
	if metadata.GetSize() <= 0 || metadata.GetSize() > limit {
		return nil, metadata, status.Error(codes.InvalidArgument, "invalid upload size")
	}
	file, err = os.CreateTemp("", "sitehub-rpc-upload-*")
	if err != nil {
		return nil, metadata, err
	}
	defer func() {
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			file = nil
		}
	}()
	var count int64
	chunk := metadata
	for {
		if err = ctx.Err(); err != nil {
			return
		}
		data := chunk.GetData()
		count += int64(len(data))
		if len(data) > 256<<10 || count > metadata.GetSize() {
			err = status.Error(codes.InvalidArgument, "upload exceeds declared size or chunk limit")
			return
		}
		if _, err = file.Write(data); err != nil {
			return
		}
		chunk, err = recv()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			return
		}
	}
	if count != metadata.GetSize() {
		err = status.Error(codes.InvalidArgument, "incomplete upload")
		return
	}
	_, err = file.Seek(0, io.SeekStart)
	return
}

func Copy(ctx context.Context, reader io.Reader, send func([]byte) error) error {
	buffer := make([]byte, 64<<10)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := reader.Read(buffer)
		if n > 0 {
			if sendErr := send(buffer[:n]); sendErr != nil {
				return sendErr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func Remove(file *os.File) {
	if file != nil {
		file.Close()
		os.Remove(file.Name())
	}
}
