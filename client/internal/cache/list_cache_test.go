package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("fetches once and serves reads from cache", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			return []string{"a", "b"}, nil
		}

		for range 5 {
			items, err := c.Get(ctx, "key", fetch)
			require.NoError(t, err)
			assert.Equal(t, []string{"a", "b"}, items)
		}
		assert.Equal(t, 1, fetches)
	})

	t.Run("keys are cached independently", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			return nil, nil
		}

		_, err := c.Get(ctx, "one", fetch)
		require.NoError(t, err)
		_, err = c.Get(ctx, "two", fetch)
		require.NoError(t, err)
		assert.Equal(t, 2, fetches)
	})

	t.Run("expired entries are refetched", func(t *testing.T) {
		c := NewListCache[string](10 * time.Millisecond)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			return []string{"a"}, nil
		}

		_, err := c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		time.Sleep(20 * time.Millisecond)
		_, err = c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		assert.Equal(t, 2, fetches)
	})

	t.Run("invalidate forces a refetch", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			return []string{"a"}, nil
		}

		_, err := c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		c.Invalidate("key")
		c.Invalidate("never-fetched") // no-op
		_, err = c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		assert.Equal(t, 2, fetches)
	})

	t.Run("invalidate all drops every key", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			return []string{"a"}, nil
		}

		_, err := c.Get(ctx, "one", fetch)
		require.NoError(t, err)
		_, err = c.Get(ctx, "two", fetch)
		require.NoError(t, err)
		c.InvalidateAll()
		_, err = c.Get(ctx, "one", fetch)
		require.NoError(t, err)
		_, err = c.Get(ctx, "two", fetch)
		require.NoError(t, err)
		assert.Equal(t, 4, fetches)
	})

	t.Run("errors are returned and not cached", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches int
		fetch := func(context.Context) ([]string, error) {
			fetches++
			if fetches == 1 {
				return nil, errors.New("transient")
			}
			return []string{"a"}, nil
		}

		_, err := c.Get(ctx, "key", fetch)
		require.Error(t, err)
		items, err := c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		assert.Equal(t, []string{"a"}, items)
		assert.Equal(t, 2, fetches)
	})

	t.Run("concurrent reads share a single fetch", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		var fetches atomic.Int64
		fetch := func(context.Context) ([]string, error) {
			fetches.Add(1)
			time.Sleep(20 * time.Millisecond) // let the other readers pile up
			return []string{"a"}, nil
		}

		var wg sync.WaitGroup
		for range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				items, err := c.Get(ctx, "key", fetch)
				assert.NoError(t, err)
				assert.Equal(t, []string{"a"}, items)
			}()
		}
		wg.Wait()
		assert.Equal(t, int64(1), fetches.Load())
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		c := NewListCache[string](time.Minute)
		fetch := func(context.Context) ([]string, error) {
			return []string{"a", "b"}, nil
		}

		first, err := c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		first[0] = "mutated"

		second, err := c.Get(ctx, "key", fetch)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, second)
	})
}
