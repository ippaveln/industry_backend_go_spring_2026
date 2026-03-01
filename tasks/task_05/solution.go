package main

type Cache[K comparable, V any] struct {
	capacity int
	inner    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		capacity: capacity,
		inner:    make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	c.inner[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		return *new(V), false
	}
	value, ok := c.inner[key]
	return value, ok
}
