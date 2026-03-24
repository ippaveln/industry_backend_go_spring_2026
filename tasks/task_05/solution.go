package main

type Cache[K comparable, V any] struct {
	cache    map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		cache:    make(map[K]V),
		capacity: capacity,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	c.cache[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		var zero V
		return zero, false
	}
	val, ok := c.cache[key]
	return val, ok
}
