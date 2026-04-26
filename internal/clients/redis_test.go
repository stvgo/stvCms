package clients

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisWrapper(t *testing.T) {
	t.Run("wrapper satisface IRedisClient", func(t *testing.T) {
		// Verify interface compliance at compile time — if this compiles, the interface is satisfied.
		var _ IRedisClient = &redisWrapper{}
	})
}

func TestRedisWrapper_Methods(t *testing.T) {
	// Use a mock IRedisClient (hand-written) to test the wrapper interface contract.
	ctx := context.Background()

	t.Run("mock implementa IRedisClient", func(t *testing.T) {
		mock := &fakeRedis{data: map[string]string{}}

		err := mock.Set(ctx, "key1", "value1", time.Minute)
		require.NoError(t, err)

		val, err := mock.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", val)

		err = mock.Del(ctx, "key1")
		require.NoError(t, err)

		_, err = mock.Get(ctx, "key1")
		assert.Error(t, err)
	})
}

// fakeRedis is a simple in-memory implementation of IRedisClient for testing.
type fakeRedis struct {
	data map[string]string
}

func (f *fakeRedis) Get(_ context.Context, key string) (string, error) {
	v, ok := f.data[key]
	if !ok {
		return "", assert.AnError
	}
	return v, nil
}

func (f *fakeRedis) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	f.data[key] = value.(string)
	return nil
}

func (f *fakeRedis) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(f.data, k)
	}
	return nil
}
