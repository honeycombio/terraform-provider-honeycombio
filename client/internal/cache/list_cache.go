// Package cache provides small read-through caches for API collections.
package cache

import (
	"context"
	"slices"
	"sync"
	"time"
)

// ListCache is a keyed, read-through cache for full-collection list results.
//
// It exists to collapse the storm of per-item lookups a large Terraform or
// Pulumi operation generates: rather than one API request per resource, the
// first read of a key fetches the full collection once and later reads are
// answered from memory. Concurrent reads of the same key share a single
// fetch, entries expire after a short TTL, and writers are expected to call
// Invalidate so a fresh list is fetched on the next read.
type ListCache[T any] struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[string]*listEntry[T]
}

type listEntry[T any] struct {
	// mu serializes fetches for this key: the first caller populates
	// while the rest wait, then everyone reads the same result.
	mu        sync.Mutex
	items     []T
	fetchedAt time.Time
	valid     bool
}

func NewListCache[T any](ttl time.Duration) *ListCache[T] {
	return &ListCache[T]{
		ttl:     ttl,
		entries: make(map[string]*listEntry[T]),
	}
}

// Get returns the collection for key, fetching it with fetch when the
// entry is missing, invalidated, or older than the TTL.
//
// The returned slice is a copy and is safe for the caller to modify.
// Fetch errors are returned to every waiting caller and are not cached.
func (c *ListCache[T]) Get(ctx context.Context, key string, fetch func(context.Context) ([]T, error)) ([]T, error) {
	c.mu.Lock()
	e, ok := c.entries[key]
	if !ok {
		e = &listEntry[T]{}
		c.entries[key] = e
	}
	c.mu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.valid || time.Since(e.fetchedAt) >= c.ttl {
		items, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		e.items = items
		e.fetchedAt = time.Now()
		e.valid = true
	}
	return slices.Clone(e.items), nil
}

// Invalidate drops the cached collection for key so the next Get fetches
// a fresh copy. Writers should call this after any mutation of the
// collection.
func (c *ListCache[T]) Invalidate(key string) {
	c.mu.Lock()
	e, ok := c.entries[key]
	c.mu.Unlock()
	if !ok {
		return
	}
	invalidate(e)
}

// InvalidateAll drops every cached collection. Writers should call this
// after a mutation which may be visible under other keys too.
func (c *ListCache[T]) InvalidateAll() {
	c.mu.Lock()
	entries := make([]*listEntry[T], 0, len(c.entries))
	for _, e := range c.entries {
		entries = append(entries, e)
	}
	c.mu.Unlock()

	for _, e := range entries {
		invalidate(e)
	}
}

func invalidate[T any](e *listEntry[T]) {
	e.mu.Lock()
	e.valid = false
	e.items = nil
	e.mu.Unlock()
}
