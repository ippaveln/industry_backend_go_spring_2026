package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]V, max(capacity, 0)),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	c.items[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V
	if c.capacity <= 0 {
		return zero, false
	}
	v, ok := c.items[key]
	if !ok {
		return zero, false
	}
	return v, true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
