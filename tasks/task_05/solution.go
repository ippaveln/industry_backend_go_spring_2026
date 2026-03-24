package main

type Cache[K comparable, V any] struct {
	enabled bool
	data    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		enabled: capacity > 0,
		data:    make(map[K]V, capacity),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if !c.enabled {
		return
	}
	c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V
	if !c.enabled {
		return zero, false
	}

	value, ok := c.data[key]
	if !ok {
		return zero, false
	}

	return value, true
}