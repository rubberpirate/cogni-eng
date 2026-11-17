package blob

import (
	"bytes"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
)

const MaxAcademicBytes = 50 << 20

func (s *Store) PutAcademic(ctx context.Context, key, mime string, data []byte) error {
	if len(data) == 0 || len(data) > MaxAcademicBytes {
		return errors.New("academic file must be between 1 byte and 50 MiB")
	}
	_, err := s.Client.PutObject(ctx, s.Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: mime, DisableMultipart: true})
	return err
}
func (s *Store) GetAcademic(ctx context.Context, key string) ([]byte, error) {
	r, err := s.Client.GetObject(ctx, s.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	data, err := io.ReadAll(io.LimitReader(r, MaxAcademicBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxAcademicBytes {
		return nil, errors.New("stored academic file exceeds limit")
	}
	return data, nil
}
