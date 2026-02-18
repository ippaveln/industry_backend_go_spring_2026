package main

type Cache[K comparable, V any] struct {
	values   map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		values: make(map[K]V, capacity),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		var v V
		return v, false
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
