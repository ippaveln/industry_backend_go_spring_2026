package main

type Cache[K comparable, V any] struct {
	store    map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{store: make(map[K]V, capacity), capacity: capacity}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}

	c.store[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		var zeroValue V
		return zeroValue, false
	}

	v, ok := c.store[key]
	return v, ok
}
