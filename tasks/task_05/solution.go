package main

type Cache[K comparable, V any] struct {
	values   map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		values: make(map[K]V, capacity),
		capacity: capacity,
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		var zero V
		return zero, false
	}
	v, ok := c.values[key]
	return v, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	c.values[key] = value
}
