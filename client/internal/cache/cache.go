// Package cache provides small read-through caches for API collections.
package cache

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
)

// Cache is a keyed, read-through cache for full-collection list results.
//
// It exists to collapse the storm of per-item lookups a large Terraform or
// Pulumi operation generates: rather than one API request per resource, the
// first read of a key fetches the full collection once and later reads are
// answered from memory. Concurrent reads of the same key share a single
// fetch, entries expire after a short TTL, and writers are expected to call
// Invalidate so a fresh list is fetched on the next read.
//
// Storage is hashicorp/golang-lru's expirable cache (unbounded, TTL-evicted)
// and fetch de-duplication is x/sync/singleflight. The one piece they don't
// provide together is invalidation ordering: an Invalidate issued while a
// fetch is in flight must not be lost when that fetch lands. The epoch
// counter below provides it — reads capture the epoch before fetching and
// only store results when no invalidation happened in between, with mu
// making the check-and-store atomic against a concurrent invalidation.
type Cache[T any] struct {
	lru   *expirable.LRU[string, []T]
	group singleflight.Group

	// epoch is bumped by every invalidation and does two jobs. It
	// qualifies the singleflight key, so a read begun after an
	// invalidation never joins (and is never served by) a fetch begun
	// before it — reads only share fetches from their own epoch, which
	// preserves read-your-writes for the invalidating goroutine. And a
	// fetched result is only stored when the epoch is unchanged since
	// before the fetch began, so an invalidation during an in-flight
	// fetch wins over the (possibly stale) fetch result. Bumping it
	// for any key is deliberately coarse: the cost of a false positive
	// is one extra list fetch on the next read, never a stale entry.
	//
	// mu serializes invalidations against the post-fetch epoch check
	// and store: without it an invalidation landing between the check
	// and the store would be lost, caching the stale result anyway.
	mu    sync.Mutex
	epoch atomic.Uint64
}

func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		// size 0 disables the LRU bound: entries are evicted by TTL
		// only. The keyspace is the handful of datasets an operation
		// touches, so it stays small.
		lru: expirable.NewLRU[string, []T](0, nil, ttl),
	}
}

// Get returns the collection for key, fetching it with fetch when the
// entry is missing, invalidated, or older than the TTL.
//
// The returned slice is a copy and is safe for the caller to modify.
// Fetch errors are returned to every caller sharing the fetch and are not
// cached. The fetch runs under the initiating caller's context; a waiter
// whose own context is cancelled stops waiting and returns its ctx error.
func (c *Cache[T]) Get(ctx context.Context, key string, fetch func(context.Context) ([]T, error)) ([]T, error) {
	epoch := c.epoch.Load()
	if items, ok := c.lru.Get(key); ok {
		return slices.Clone(items), nil
	}

	// the flight key is epoch-qualified so this read can only share a
	// fetch begun in its own epoch: a fetch made stale by an
	// invalidation still completes for the readers that predate the
	// invalidation, but readers arriving after it start a fresh fetch
	ch := c.group.DoChan(fmt.Sprintf("%d\x00%s", epoch, key), func() (any, error) {
		items, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		if c.epoch.Load() == epoch {
			c.lru.Add(key, items)
		}
		c.mu.Unlock()
		return items, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		items, ok := res.Val.([]T)
		if !ok {
			// unreachable: the flight's function above always returns a []T
			return nil, fmt.Errorf("cache: unexpected singleflight result type %T", res.Val)
		}
		return slices.Clone(items), nil
	}
}

// Invalidate drops the cached collection for key so the next Get fetches
// a fresh copy. Writers should call this after any mutation of the
// collection.
func (c *Cache[T]) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.epoch.Add(1)
	c.lru.Remove(key)
}

// InvalidateAll drops every cached collection. Writers should call this
// after a mutation which may be visible under other keys too.
func (c *Cache[T]) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.epoch.Add(1)
	c.lru.Purge()
}
