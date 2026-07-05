// Package mediastore is the infrastructure adapter that stores incident images
// in an S3-compatible object storage (e.g. Selectel). It implements the domain
// media.Storage port and uploads objects with a public-read ACL so the
// returned URLs can be embedded directly by the frontend.
package mediastore

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

type Config struct {
	EndpointURL   string
	Region        string
	Bucket        string
	AccessKey     string
	SecretKey     string
	PublicBaseURL string
}

// Storage uploads images to an S3 bucket. It implements media.Storage.
type Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func New(cfg Config) *Storage {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.EndpointURL),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: true,
	})

	return &Storage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

func (s *Storage) Upload(ctx context.Context, key string, file media.File) (string, error) {
	const op = "mediastore.Upload"

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file.Data),
		ContentType: aws.String(file.ContentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", errs.Wrapf(errs.KindUnavailable, op, err, "put object")
	}

	return s.publicBaseURL + "/" + key, nil
}
