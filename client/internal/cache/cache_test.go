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

func TestCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("fetches once and serves reads from cache", func(t *testing.T) {
		c := New[string](time.Minute)
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
		c := New[string](time.Minute)
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
		c := New[string](10 * time.Millisecond)
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
		c := New[string](time.Minute)
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
		c := New[string](time.Minute)
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
		c := New[string](time.Minute)
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
		c := New[string](time.Minute)
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

	t.Run("invalidation during an in-flight fetch is not lost", func(t *testing.T) {
		c := New[string](time.Minute)
		var fetches atomic.Int64
		fetchStarted := make(chan struct{})
		releaseFetch := make(chan struct{})

		firstRead := make(chan struct{})
		go func() {
			defer close(firstRead)
			_, _ = c.Get(ctx, "key", func(context.Context) ([]string, error) {
				fetches.Add(1)
				close(fetchStarted)
				<-releaseFetch
				return []string{"stale"}, nil
			})
		}()

		// invalidate while the fetch is in flight: its result must not
		// be stored, or the invalidation would be silently lost
		<-fetchStarted
		c.Invalidate("key")
		close(releaseFetch)
		<-firstRead

		items, err := c.Get(ctx, "key", func(context.Context) ([]string, error) {
			fetches.Add(1)
			return []string{"fresh"}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"fresh"}, items, "expected a fresh fetch, not the pre-invalidation result")
		assert.Equal(t, int64(2), fetches.Load())
	})

	t.Run("reads after an invalidation do not join a pre-invalidation fetch", func(t *testing.T) {
		c := New[string](time.Minute)
		fetchStarted := make(chan struct{})
		releaseFetch := make(chan struct{})

		staleRead := make(chan []string, 1)
		go func() {
			items, _ := c.Get(ctx, "key", func(context.Context) ([]string, error) {
				close(fetchStarted)
				<-releaseFetch
				return []string{"stale"}, nil
			})
			staleRead <- items
		}()

		// a read arriving after the invalidation must start a fresh
		// fetch, not be served by the still-in-flight pre-invalidation
		// one — that would hand a post-write reader pre-write data
		<-fetchStarted
		c.Invalidate("key")

		items, err := c.Get(ctx, "key", func(context.Context) ([]string, error) {
			return []string{"fresh"}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"fresh"}, items)

		// the doomed fetch still completes and serves the reader which
		// started before the invalidation
		close(releaseFetch)
		assert.Equal(t, []string{"stale"}, <-staleRead)

		// and its result was discarded, not cached: the fresh entry
		// survives without a refetch
		var refetched bool
		items, err = c.Get(ctx, "key", func(context.Context) ([]string, error) {
			refetched = true
			return nil, nil
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"fresh"}, items)
		assert.False(t, refetched, "expected the fresh entry to still be cached")
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		c := New[string](time.Minute)
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
