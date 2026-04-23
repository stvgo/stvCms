package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewR2Client(ctx context.Context, accountID, accessKeyID, secretAccessKey string) *s3.Client {
	cfg, _ := config.LoadDefaultConfig(ctx,
		config.WithBaseEndpoint(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)

	return s3.NewFromConfig(cfg)
}
