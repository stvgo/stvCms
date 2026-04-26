package clients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate mockgen -destination=../mocks/mock_r2_client.go -package=mocks stvCms/internal/clients IR2Client

type IR2Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
