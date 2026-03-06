package main

type Cache[K comparable, V any] struct {
	data map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		return &Cache[K, V]{}
	}
	return &Cache[K, V]{
		data: make(map[K]V, capacity),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.data != nil {
		c.data[key] = value
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.data == nil {
		var zero V
		return zero, false
	}

	val, ok := c.data[key]
	return val, ok
}
