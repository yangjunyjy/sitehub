package streamio

import (
	"context"
	"fmt"
	"io"
)

type Reader struct {
	Next   func() ([]byte, error)
	Cancel context.CancelFunc
	buffer []byte
}

func (r *Reader) Read(p []byte) (int, error) {
	for len(r.buffer) == 0 {
		data, err := r.Next()
		if err != nil {
			return 0, err
		}
		r.buffer = data
	}
	n := copy(p, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}
func (r *Reader) Close() error {
	if r.Cancel != nil {
		r.Cancel()
	}
	return nil
}

// Seeker opens a fresh bounded RPC stream after a seek. ServeContent can therefore
// implement HEAD, single/multipart ranges and content length without buffering.
type Seeker struct {
	Size   int64
	Open   func(offset, length int64) (io.ReadCloser, error)
	offset int64
	body   io.ReadCloser
}

func (s *Seeker) Read(p []byte) (int, error) {
	if s.offset >= s.Size {
		return 0, io.EOF
	}
	if s.body == nil {
		body, err := s.Open(s.offset, s.Size-s.offset)
		if err != nil {
			return 0, err
		}
		s.body = body
	}
	n, err := s.body.Read(p)
	s.offset += int64(n)
	return n, err
}
func (s *Seeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += s.offset
	case io.SeekEnd:
		offset += s.Size
	default:
		return 0, fmt.Errorf("invalid seek origin")
	}
	if offset < 0 {
		return 0, fmt.Errorf("negative seek")
	}
	s.Close()
	s.offset = offset
	return offset, nil
}
func (s *Seeker) Close() error {
	if s.body != nil {
		err := s.body.Close()
		s.body = nil
		return err
	}
	return nil
}
