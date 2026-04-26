package clients

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewR2Client(t *testing.T) {
	t.Run("crea cliente con credenciales", func(t *testing.T) {
		ctx := context.Background()
		client := NewR2Client(ctx, "account123", "key123", "secret123")
		require.NotNil(t, client)
		var _ IR2Client = client
	})

	t.Run("IR2Client interface compliance con stub", func(t *testing.T) {
		var _ IR2Client = &fakeR2{}
	})
}

func TestIR2ClientInterface(t *testing.T) {
	ctx := context.Background()
	client := NewR2Client(ctx, "acc", "key", "secret")
	assert.NotNil(t, client)
}

// fakeR2 is a stub that correctly implements IR2Client.
type fakeR2 struct{}

func (f *fakeR2) PutObject(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeR2) GetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{}, nil
}
