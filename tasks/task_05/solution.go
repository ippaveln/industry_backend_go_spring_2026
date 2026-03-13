package main

type Cache[K comparable, V any] struct {
	matching map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity == 0 {
		return &Cache[K, V]{}
	}
	return &Cache[K, V]{make(map[K]V, capacity)}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.matching == nil {
		return
	}
	c.matching[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.matching == nil {
		var val V
		return val, false
	}
	value, ok := c.matching[key]
	return value, ok
}
