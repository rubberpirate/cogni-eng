package blob

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/url"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const MaxAvatarBytes = 4 << 20

type Store struct {
	Client *minio.Client
	Bucket string
}

func FromEnv() (*Store, error) {
	endpoint := os.Getenv("OBJECT_ENDPOINT")
	if endpoint == "" {
		return nil, nil
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || u.Path != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("OBJECT_ENDPOINT must be an origin URL")
	}
	if u.Scheme != "https" && (os.Getenv("APP_ENV") == "production" || u.Scheme != "http") {
		return nil, errors.New("object storage requires HTTPS in production")
	}
	key, secret, bucket := os.Getenv("OBJECT_USER"), os.Getenv("OBJECT_PASSWORD"), os.Getenv("OBJECT_BUCKET")
	if key == "" || secret == "" || bucket == "" {
		return nil, errors.New("object storage credentials and bucket are required")
	}
	client, err := minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(key, secret, ""), Secure: u.Scheme == "https", Region: "us-east-1"})
	if err != nil {
		return nil, err
	}
	return &Store{client, bucket}, nil
}
func NormalizeAvatar(data []byte) ([]byte, error) {
	if len(data) > MaxAvatarBytes {
		return nil, errors.New("image exceeds 4 MiB")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != "png" && format != "jpeg" {
		return nil, errors.New("use a PNG or JPEG image")
	}
	if cfg.Width < 32 || cfg.Width != cfg.Height || cfg.Width > 1024 {
		return nil, errors.New("crop to a square image between 32 and 1024 pixels")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("image could not be decoded")
	}
	var out bytes.Buffer
	if err = png.Encode(&out, img); err != nil {
		return nil, err
	}
	if out.Len() > MaxAvatarBytes {
		return nil, errors.New("normalized image exceeds 4 MiB; use a smaller crop")
	}
	return out.Bytes(), nil
}
func (s *Store) Put(ctx context.Context, key string, data []byte) error {
	_, err := s.Client.PutObject(ctx, s.Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/png", DisableMultipart: true})
	return err
}
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.Client.GetObject(ctx, s.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		// Read errors are returned below; closing this read-only stream only releases resources.
		_ = reader.Close()
	}()
	data, err := io.ReadAll(io.LimitReader(reader, MaxAvatarBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxAvatarBytes {
		return nil, errors.New("stored image exceeds limit")
	}
	return data, nil
}
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.Client.RemoveObject(ctx, s.Bucket, key, minio.RemoveObjectOptions{})
}
